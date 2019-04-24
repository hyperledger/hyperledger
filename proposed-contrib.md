There were six initial proposed contributions to the Hyperledger project. They are listed here for posterity.

For information on the currently active projects, please visit the [README](README.md)

## Ripple Proposed Contribution

Rippled is an open source, actively maintained, C++ implementation of a public, distributed ledger. The Ripple network has been operational as a public network since 2012, providing cross-currency atomic payments. Rippled includes a high-performance group of core classes for resistance to algorithmic complexity attacks, resource management and binary representations of ledgers and transactions. Because Ripple's use case includes complex, cross-currency payments using order books, there is significant infrastructure to simplify the development of transactors that implement complex semantics.

[https://github.com/ripple/rippled](https://github.com/ripple/rippled)

Rippled includes NuDB, a high-performance, scalable key/value store specifically designed for distributed ledger applications. NuDB is optimized for handling data sets of many terabytes with minimal RAM consumption. It requires a high-performance I/O back-end, such as an SSD.

[https://github.com/ripple/rippled/tree/develop/src/beast/beast/nudb](https://github.com/ripple/rippled/tree/develop/src/beast/beast/nudb)

## IBM Proposed Contribution

Open Blockchain (OBC) is IBM's proposed contribution to the Hyperledger project. It is a low level blockchain fabric that has been designed to meet the requirements of a variety of industry-focused use cases. It extends the learning of the pioneers in this field by addressing additional requirements needed to satisfy those broader industry use cases. The central elements of this implementation are smart contracts (what IBM calls chain code), digital assets, record repositories, a decentralized network providing consensus, and cryptographic security. To these blockchain staples, the implementation supports key industry requirements such as performance, verified identities, private and confidential transactions. Finally, the fabric is architected to provide for a pluggable consensus model, allowing a variety of specialized or optimized consensus algorithms to be applied.

[https://github.com/openblockchain](https://github.com/openblockchain)

## Digital Asset Holdings Proposed Contribution

Digital Asset's Hyperledger candidate contribution is an enterprise ready blockchain server with a client API. HLP-Candidate has a modular architecture and configurable network architecture particularly designed to meet the needs of our financial services clients. HLP-Candidate implements an append-only log of financial transactions designed to be replicated at multiple organizations without centralized control. The goal of HLP-Candidate is to allow expansion of the data backbone concept to the multi-organization level. We are open sourcing this project with the belief that as a critical part of the new financial infrastructure, this part of the software stack should be commoditized, collaborated on and serve as the robust backbone on which applications can be developed.

[https://github.com/DigitalAssetCom/hlp-candidate](https://github.com/DigitalAssetCom/hlp-candidate)

## Blockstream Proposed Contribution

Blockstream is contributing [the Elements Project](https://elementsproject.org), a modularized fork of the Bitcoin codebase that adds several major improvements called "Elements".  Elements are composable features that allow a blockchain's attributes to be customized, including [Confidential Transactions](https://elementsproject.org/elements/confidential-transactions), [Segregated Witness](https://elementsproject.org/elements/segregated-witness), and [Deterministic Pegs](https://elementsproject.org/elements/deterministic-pegs).  Sidechains are interoperable blockchains implementing atomic, cross-chain transactions using a choice of federated, permissioned, or decentralized consensus models.  This model allows HyperLedger to interoperate with the existing developer community – sharing progress on testing, scalability, and features – by allowing anyone in the world to utilize shared infrastructure to solve domain-specific problems on purpose-built sidechains.

[https://github.com/ElementsProject/elements](https://github.com/ElementsProject/elements)

## Intel's Sawtooth Lake

Designed for versatility and scalability, [Sawtooth Lake](http://intelledger.github.io/) is Intel’s modular blockchain suite.  Distributed Ledger Technology has potential in many fields with use cases from IoT to Financials.  This architecture recognizes the diversity of requirements across that spectrum.  Sawtooth Lake supports both permissioned and permissionless deployments.  It includes a novel consensus algorithm, Proof of Elapsed Time (PoET).  PoET targets large distributed validator populations with minimal resource consumption.  Transaction business logic is decoupled from the consensus layer into Transaction Families that allow for restricted or unfettered semantics.

[https://github.com/intelledger]([https://github.com/intelledger])

## Defunct Hyperledger Incubator

The LF had initially setup an incubator org for various proposals and prototyping efforts. However, this hyperledger-incubator is no longer used.
