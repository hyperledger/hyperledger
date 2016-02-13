# Quantumledger Project
Quantumledger Project is a proposal for a new kind of ledger to be used by the Hyperledger Project as a starting point. At the moment, it only specifies an API through example usages. No real implementation. 

## Principle of least power

Hyperledger is built around the principle of least power. The data structure with the least power imaginable is a POJO. It is understood by any developer and almost any platform and more-or-less interchangable with any JSON structure and can be understood by any modern JSON-based API.

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
    
