# Quantumledger Project
Quantumledger Project is a proposal for a new kind of ledger to be used by the Hyperledger Project as a starting point. At the moment, it only specifies an API through example usages. No real implementation. 

## Proof of Trust

Most crypto-currency networks derive value of the information they contain by showing some sort of "proof". For example, bitcoin uses a "proof of work" approach and derives value from the work required to search for hashes.

Quantumledger uses a new mechanism called "Proof of Trust". In its essence, it is very simple. It means deriving value from trust. Trust in this context means trust that the information provided is correct. The question is "how does one proof trust"? This can be answered by human nature. If we acquire a piece of information that we trust, we feel confident and comfortable sharing that piece of information. By sharing a piece of information we therefor have proven "trust" in this information. Since false information (also known as "lies") are mostly created arbitrarily by individuals, they will appear as outliers when asking the network for information and are overwhelmed by information the majority of the network has proven trust in. This way it is easy to distinguish a lie (false information) from true information.

## Principle of least power

Quantumledger is built around the [principle of least power](http://blog.codinghorror.com/the-principle-of-least-power/). The data structure with the least power imaginable is a POJO. It is understood by any developer and almost any platform and more-or-less interchangable with any JSON structure and can be understood by any modern JSON-based API.

## Simple query API

    // query the network for any arbitrary value
    network.ask('citybank.accounts'); // get a list of all accounts at citybank (just an example)
    network.ask('citybank.accounts.gS93Kfas3Fkwe3.balances.USD');

## Participate in the Network

    // Initialize a new ledger (it's just a POJO)
    var ledger = {}; 
    
    // Add information to the ledger
    var ledger['name'] = 'Stefan';
    
    // Create a new network that only knows about the information I added myself
    var network = Hyperledger.createNetwork(ledger); 
    
    // Include information of a node that I trust via its IP address
    var otherNode = network.include('240f:11:9b79:1:91af:3f1b:d463:8f23')
    
    // Query the network for information
    network.ask('name').listen(function(node, answer) {
      console.log(node + ' gave the answer: ' + answer);
    });
    
    // Deciding the "true" response: It's up to the developer. Recommended is to wait a little bit and then take the majority as the true answer
    // For the query about the information 'name', we can assume that any return value is true, but we don't really care about that information.
    // More interesting would be a query like `network.ask('morganstanly.accounts.fe3za9f9e9as.balances.USD')`
    
## Implementation & Networking

Quantumledger transmits information via simple HTTPS. HTTPS is the most widely used networking protocol to transfer information in a secure way without risking a man-in-the-middle attack and has now power on its own. An implementation would simply make a request to `https://[ip-address]/the.information.im.asking.for.json`. The data (JSON stream) returned can potentially be extremly large. Therefore the implementation should be able to process the JSON stream even though it is not complete yet in order to determine if enough information has been received and the request can be terminated. We can - for the start - try to use a streaming library like oboe.js. Although in the future we should implement our own solution.
