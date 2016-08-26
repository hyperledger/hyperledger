# Hyperledger Quick-Start Guide #
## What is the Hyperledger Project? ##
“The Hyperledger Project is a collaborative effort created to advance Blockchain technology by identifying and addressing important features for a cross-industry open standard for distributed ledgers that can transform the way business transactions are conducted globally.”

## Platforms ##
### Sawtooth Lake ###
“Sawtooth Lake” is a highly modular platform for building, deploying and running distributed ledgers. It features transaction families that can be customized for various use cases and applications.	
### Fabric ###
Fabric leverages familiar and proven technologies featuring container technology for smart contract development.  Its modular architecture allows pluggable implementations of various functions. 

## Which One is right for me? ##
The inherent nature of distributed ledgers requires tradeoffs between performance capabilities. Below is a chart to help determine which platform is most appropriate for your application. 

### Sawtooth Lake	 ###

**Consensus Mechanisms:**	

1) PoET: Proof of Elapsed Time.  A form of random leader election using trusted execution in lieu of mining rigs.

2) Quorum: A form of PBFT using intersecting subsets of the graph.
	
**Language:**	Python

**Network Type(s):** Permissioned & permissionless

**Max Network Size:** ~100

**Use Case Development:**  	

Transaction Families: Transaction logic installed on the validator to maximize efficiency. Examples provided as extensions such as MarketPlace and Arcade. 

**Deployment:**	Any cloud provider or locally on premises.
 


### Fabric ##

**Consensus Mechanisms:** 

1) No-op (consensus ignored); 

2) Batch PBFT

**Language:** Go 

**Network Type(s):** Permissioned 

**Max Network Size:** ~6

**Use Case Development:**  	Chain code. Docker deployed logic submitted as transactions to maximize deployment ease of use and upgradability.

**Deployment:**	IBM cloud services
	

----------
## Get involved ##
[Sawtooth Lake Tutuorial ](http://intelledger.github.io/tutorial.html "Sawtooth Lake GitHub")

[Fabric docs
](http://hyperledger-fabric.readthedocs.io/en/latest/)


