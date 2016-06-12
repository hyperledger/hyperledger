# Hyperledger Project

Thank you for your interest in the Hyperledger Project, a Collaborative Project at the Linux Foundation. The Hyperledger Project aims to develop a distributed ledger platform that can provide building blocks for a wide range of use cases and value-add solutions. The project operates with some basic principles:

* Given the diversity of usages, modularity is critical. Examples include transaction semantics, contract languages, consensus, identity, and storage.

* Code speaks. The objective of the project is to develop technologies that can be used to build and deploy distributed ledgers quickly and easily.

* The project will evolve over time with better understanding of requirements and emerging usages. While the objective is to develop a single platform, we expect that platform to emerge from a diverse set of approaches.

Check [the community page](http://hyperledger.org/community) for more information about joining the mailing lists and participating in the conversations. Contributions to the Hyperledger Project are expected to progress through a standard [life cycle](https://github.com/hyperledger/hyperledger/wiki/Project-Lifecycle) from incubation through maturity. Several codebases have been offered as a starting point for evaluation. There are currently two top-level projects under incubation.

## Contributions most welcome!
We invite anyone with an interest to participate via contributions to any and all of the Hyperledger's projects. If you are interested in tracking our progress, we document and record all decisions and project proposals etc on our [wiki](https://github.com/hyperledger/hyperledger/wiki). There, you'll also find a few working groups working on such activities as collecting and documenting [use cases and requirements](https://github.com/hyperledger/hyperledger/wiki/Requirements-WG), discussing and developing the architecture, exploring issues and solutions for dealing with the challenges of Identity and working on articulting our overall vision through development of a [whitepaper](https://github.com/hyperledger/hyperledger/wiki/Whitepaper-WG). Each of these activities, as well as hacking on the various projects, listed below, are open to all that wish to engage.

## Fabric Incubator

[Fabric](https://github.com/hyperledger/fabric) is a project in [Incubation](https://github.com/hyperledger/hyperledger/wiki/Project-Lifecycle) that was proposed by Tamas Blummer (DAH) and Christopher Ferris (IBM) as a result of the first hackathon during which a merge between IBM's proposal and DAH's proposal was started (see [Proposal](https://docs.google.com/document/d/1XECRVN9hXGrjAjysrnuNSdggzAKYm6XESR6KmABwhkE)).

The fabric is an implementation of blockchain technology that is intended as a foundation for developing blockchain applications or solutions. It offers a modular architecture allowing components, such as consensus and membership services, to be plug-and-play. It leverages container technology to host smart contracts called "chaincode" that comprise the application logic of the system. 

There are two repositories that comprise the fabric incubator:
[https://github.com/hyperledger/fabric](https://github.com/hyperledger/fabric)
[https://github.com/hyperledger/fabric-api](https://github.com/hyperledger/fabric-api)

You can find out how to contribute to the Fabric project [here](https://github.com/hyperledger/fabric/blob/master/CONTRIBUTING.md) and [here](https://github.com/hyperledger/fabric-api/blob/master/docs/contributing.md).

## Sawtooth Lake Incubator

Designed for versatility and scalability, [Sawtooth Lake](http://hyperledger/sawtooth-core) is Intel’s modular blockchain suite.  Distributed Ledger Technology has potential in many fields with use cases from IoT to Financials.  This architecture recognizes the diversity of requirements across that spectrum.  Sawtooth Lake supports both permissioned and permissionless deployments.  It includes a novel consensus algorithm, Proof of Elapsed Time (PoET).  PoET targets large distributed validator populations with minimal resource consumption.  Transaction business logic is decoupled from the consensus layer into Transaction Families that allow for restricted or unfettered semantics.

There are six repositories that comprise the Sawtooth Lake project:
[https://github.com/hyperledger/sawtooth-core](https://github.com/hyperledger/sawtooth-core)
[https://github.com/hyperledger/sawtooth-validator](https://github.com/hyperledger/sawtooth-validator)
[https://github.com/hyperledger/sawtooth-arcade](https://github.com/hyperledger/sawtooth-arcade)
[https://github.com/hyperledger/sawtooth-docs](https://github.com/hyperledger/sawtooth-docs)
[https://github.com/hyperledger/sawtooth-mktplace](https://github.com/hyperledger/sawtooth-mktplace)
[https://github.com/hyperledger/sawtooth-dev-tools](https://github.com/hyperledger/sawtooth-dev-tools)

You can find out how to contribute to the Sawtooth lake project [here](https://github.com/hyperledger/sawtooth-core/blob/master/CONTRIBUTING.md)
