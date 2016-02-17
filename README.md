# Hyperledger Project
Hyperledger Project is a new Collaborative Project at The Linux Foundation. The technical community is just getting started and will be adding code to the repository in the coming weeks. Check hyperledger.org for more information about joining the mailing lists and participating in the conversations.

Thank you for your interest in the Hyperledger Project. Currently there have been four proposed contributions with potentially others in the future. At present the TSC and community are looking at ways to build a platform that can provide common plumbing for a wide range of use cases and value added solutions on top. Please keep in mind all of the proposals below are simply proposals and that the community will evaluate various ways to get started. Below are the links to the codebases for evaluation purposes, in no particular order:

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

Blockstream proposes to contribute its Elements codebase available at [https://github.com/ElementsProject/elements](https://github.com/ElementsProject/elements)
