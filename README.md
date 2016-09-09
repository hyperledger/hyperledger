# Hyperledger Project

Thank you for your interest in the Hyperledger Project, a Collaborative Project at the Linux Foundation. The Hyperledger Project aims to develop a distributed ledger platform that can provide building blocks for a wide range of use cases and value-add solutions. The project operates with some basic principles:

* Given the diversity of usages, modularity is critical. Examples include transaction semantics, contract languages, consensus, identity, and storage.

* Code speaks. The objective of the project is to develop technologies that can be used to build and deploy distributed ledgers quickly and easily.

* The project will evolve over time with better understanding of requirements and emerging usages. While the objective is to develop a single platform, we expect that platform to emerge from a diverse set of approaches.

Check [the community page](http://hyperledger.org/community) for more information about joining the mailing lists and participating in the conversations. Contributions to the Hyperledger Project are expected to progress through a standard [life cycle](https://github.com/hyperledger/hyperledger/wiki/Project-Lifecycle) from incubation through maturity. Several codebases have been offered as a starting point for evaluation. There are currently two top-level projects under incubation.

## Contributions most welcome!
We invite anyone with an interest to participate via contributions to any and all of the Hyperledger's projects. If you are interested in tracking our progress, we document and record all decisions and project proposals etc on our [wiki](https://github.com/hyperledger/hyperledger/wiki). There, you'll also find a few working groups working on such activities as collecting and documenting [use cases and requirements](https://github.com/hyperledger/hyperledger/wiki/Requirements-WG), discussing and developing the architecture, exploring issues and solutions for dealing with the challenges of Identity and working on articulting our overall vision through development of a [whitepaper](https://github.com/hyperledger/hyperledger/wiki/Whitepaper-WG). Each of these activities, as well as hacking on the various projects, listed below, are open to all that wish to engage.

## Blockchain Explorer Incubator
[Blockchain-explorer](https://gerrit.hyperledger.org/r/gitweb?p=blockchain-explorer.git;a=tree) is a project in [Incubation](https://github.com/hyperledger/hyperledger/wiki/Project-Lifecycle) that was proposed by Christopher Ferris (IBM), Dan Middleton (Intel) and Pardha Vishnumolakala (DTCC) to create a user friendly web application for Hyperledger to view/query blocks, transactions and associated data, network information (name, status, list of nodes), chain codes/transaction families (view/invoke/deploy/query) and any other relevant information stored in the ledger.

## Fabric Incubator

The Hyperledger [fabric](https://github.com/hyperledger/fabric) is a project in [Incubation](https://github.com/hyperledger/hyperledger/wiki/Project-Lifecycle) that was proposed by Tamas Blummer (DAH) and Christopher Ferris (IBM) as a result of the first hackathon during which a merge between IBM's proposal and DAH's proposal was started (see [Proposal](https://docs.google.com/document/d/1XECRVN9hXGrjAjysrnuNSdggzAKYm6XESR6KmABwhkE)).

The fabric is an implementation of blockchain technology that is intended as a foundation for developing blockchain applications or solutions. It offers a modular architecture allowing components, such as consensus and membership services, to be plug-and-play. It leverages container technology to host smart contracts called "chaincode" that comprise the application logic of the system. 

There are three repositories that comprise the fabric incubator:

* [fabric (Gerrit)](https://gerrit.hyperledger.org/r/gitweb?p=fabric.git;a=tree)
* [fabric-api (Gerrit)](https://gerrit.hyperledger.org/r/gitweb?p=fabric-api.git;a=tree)
* [fabric-chaintool](https://github.com/hyperledger/fabric-chaintool)

**Note:** we also maintain _read-only_ mirrors of our Gerrit managed repos on GitHub:

* [fabric](https://github.com/hyperledger/fabric)
* [fabric-api](https://github.com/hyperledger/fabric-api)

You can find out how to contribute to the fabric project [here](http://hyperledger-fabric.readthedocs.io/en/latest/CONTRIBUTING/). Complete documentation for the fabric can be found [here](http://hyperledger-fabric.readthedocs.io/en/latest/). We use [Jira](https://jira.hyperledger.org/secure/RapidBoard.jspa?rapidView=7&view=planning.nodetail) for tracking issues and development.

## Sawtooth Lake Incubator

Designed for versatility and scalability, [Sawtooth Lake](https://github.com/hyperledger/sawtooth-core) is Intel’s modular blockchain suite.  Distributed Ledger Technology has potential in many fields with use cases from IoT to Financials.  This architecture recognizes the diversity of requirements across that spectrum.  Sawtooth Lake supports both permissioned and permissionless deployments.  It includes a novel consensus algorithm, Proof of Elapsed Time (PoET).  PoET targets large distributed validator populations with minimal resource consumption.  Transaction business logic is decoupled from the consensus layer into Transaction Families that allow for restricted or unfettered semantics.

The Sawtooth Lake project is contained in a single repository. The tools and code that were previously contained in multiple repositories have been consolidated:

* [https://github.com/hyperledger/sawtooth-core](https://github.com/hyperledger/sawtooth-core)

You can find out how to contribute to the Sawtooth lake project [here](https://github.com/hyperledger/sawtooth-core/blob/master/CONTRIBUTING.md)
