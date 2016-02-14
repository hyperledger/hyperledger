# Quantumledger Project
Quantumledger Project is a proposal for a new kind of ledger to be used by the Hyperledger Project as a starting point. At the moment, it only specifies an API through example usages. No real implementation. 

## Proof of Trust

Most crypto-currency networks derive value of the information they contain by showing some sort of "proof". For example, bitcoin uses a "proof of work" approach and derives value from the work required to search for hashes.

Quantumledger uses a new mechanism called "Proof of Trust". In its essence, it is very simple. It means deriving value from trust. Trust in this context means trust that the information provided is correct. The question is "how does one proof trust"? This can be answered by human nature. If we acquire a piece of information that we trust, we feel confident and comfortable sharing that piece of information. By sharing a piece of information we therefor have proven "trust" in this information. Since false information (also known as "lies") are mostly created arbitrarily by individuals, they will appear as outliers when asking the network for information and are overwhelmed by information the majority of the network has proven trust in. This way it is easy to distinguish a lie (false information) from true information.

### Chain of Trust, example

Individual `alice` learns information `A`. `alice` believes that `A` is true, therefor he trusts that information. He adds `A` to his ledger and shares this information. Individual `bob` asks `alice` for information. `alice` shares information `A` with `bob`. By sharing `A` with `bob`, `alice` has demonstrated trust in `A`. `alice` has shown proof of trust in `A`. `bob` can now decide if he thinks `A` is true or not. If `bob` trusts `alice`, he should be able to conclude that `A` is true. `bob` is now able to proof trust by sharing `A` among the network, joining the chain of trust.

## Principle of least power

Quantumledger is built around the [principle of least power](http://blog.codinghorror.com/the-principle-of-least-power/). The data structure with the least power imaginable is a POJO. It is understood by any developer and almost any platform and more-or-less interchangable with any JSON structure and can be understood by any modern JSON-based API.

## Simple query API

    // query the network for any arbitrary value
    network.ask('citybank/accounts'); // get a list of all accounts at citybank (just an example)
    network.ask('citybank/accounts/gS93Kfas3Fkwe3/balances/USD');
    
    // download all information stored across the known network
    network.ask('**/*');

## Participate in the Network

    // Initialize a new ledger (it's just a POJO)
    var ledger = {}; 
    
    // Add information to the ledger
    ledger['name'] = 'Stefan';
    
    // Create a new network that only knows about the information I added myself
    var network = Hyperledger.createNetwork(ledger); 
    
    // Include information of a node that I trust via its IP address and public key
    var otherNode = network.include('240f:11:9b79:1:91af:3f1b:d463:8f23/[public key]')
    
    // Ask the network for information
    network.ask('name').listen(function(node, answer) {
      console.log(node + ' gave the answer: ' + answer);
    });
    
    // Ask a node for information
    otherNode.ask('name').listen(function(node, answer) {
      console.log(node + ' gave the answer: ' + answer);
    });
    
    
### Deciding whether information is true or not

Answering this question is up to the developer. The example implementation simply waits a little while while it collects answers and then stores the majority response in its ledger.
However, this will NOT work all cases. Some information has only one true source of origin. For example, the amount of money in a bank account. The developer will have to discard answers from nodes that are not the bank, or simply not even ask those nodes.
However, some information, like `name` are only true within the boundaries of one node. Thus, it does not make sense to ask the network for the information `name` and then store this information within our own ledger.
More interesting would be a query like `network.ask('morganstanly/accounts/fe3za9f9e9as/balances/USD')`

## Node Discovery

Just ask the network for more addresses

    // return a list of more addresses
    network.ask('trusted_addresses');

## Ledger Structure & Asking

Examples

    var ledger = {
      users:  {
        fsd3fsd: {
          name: 'Peter',
          social_security_number: 234234234
        },
        jejf3jf: {
          name: 'Stefan,
          social_security_number: 123455
        }
      },
      name: 'root'
    }
    
    ask('name'); // => "root"
    ask('users'); // => ["fsd3fsd", "jejf3jf"]
    ask('users/*'); // => {"fsd3fsd": ["name", "social_security_number"], "jejf3jf": ["name", "social_security_number"]}
    ask('users/**/*'); // => {"fsd3fsd":{"name": "Peter","social_security_number": 234234234},"jejf3jf":{"name": "Stefan","social_security_number": 123455}}
    ask('users/fsd3fsd/name'); // => "Peter"
    
## Transaction

To make a transaction, all you have to do is change a value in your ledger.

    ledger.stefan.money += 50
    ledger.peter.money -= 50
    
The next time somebody asks you for this information, you give them the updated information. If you have shown enough proof of trust in the past, people will assume that the updated information is correct.

## Implementation & Networking

Quantumledger transmits information via simple HTTP. HTTP is the most widely used networking protocol to transfer information. An implementation would simply make a request to `http://[ip-address]/[publickey]/the/information/im/asking/for`. The data (JSON stream) returned can potentially be extremly large. Therefore the implementation should be able to process the JSON stream even though it is not complete yet in order to determine if enough information has been received and the request can be terminated. We can - for the start - try to use a streaming library like oboe.js. Although in the future we should implement our own solution.
