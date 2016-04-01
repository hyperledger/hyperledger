// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package rpc

var (
	// Holds geth specific RPC extends which can be used to extend web3
	WEB3Extensions = map[string]string{
		"personal": Personal_JS,
		"txpool":   TxPool_JS,
		"admin":    Admin_JS,
		"eth":      Eth_JS,
		"miner":    Miner_JS,
		"debug":    Debug_JS,
		"net":      Net_JS,
	}
)

const Personal_JS = `
web3._extend({
	property: 'personal',
	methods:
	[
		new web3._extend.Method({
			name: 'newAccount',
			call: 'personal_newAccount',
			params: 1,
			outputFormatter: web3._extend.utils.toAddress
		}),
		new web3._extend.Method({
			name: 'unlockAccount',
			call: 'personal_unlockAccount',
			params: 3,
		}),
		new web3._extend.Method({
			name: 'lockAccount',
			call: 'personal_lockAccount',
			params: 1
		})
	],
	properties:
	[
		new web3._extend.Property({
			name: 'listAccounts',
			getter: 'personal_listAccounts'
		})
	]
});
`

const TxPool_JS = `
web3._extend({
	property: 'txpool',
	methods:
	[
	],
	properties:
	[
		new web3._extend.Property({
			name: 'content',
			getter: 'txpool_content'
		}),
		new web3._extend.Property({
			name: 'inspect',
			getter: 'txpool_inspect'
		}),
		new web3._extend.Property({
			name: 'status',
			getter: 'txpool_status',
			outputFormatter: function(status) {
				status.pending = web3._extend.utils.toDecimal(status.pending);
				status.queued = web3._extend.utils.toDecimal(status.queued);
				return status;
			}
		})
	]
});
`

const Admin_JS = `
web3._extend({
	property: 'admin',
	methods:
	[
		new web3._extend.Method({
			name: 'addPeer',
			call: 'admin_addPeer',
			params: 1
		}),
		new web3._extend.Method({
			name: 'exportChain',
			call: 'admin_exportChain',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'importChain',
			call: 'admin_importChain',
			params: 1
		}),
		new web3._extend.Method({
			name: 'sleepBlocks',
			call: 'admin_sleepBlocks',
			params: 2
		}),
		new web3._extend.Method({
			name: 'setSolc',
			call: 'admin_setSolc',
			params: 1
		}),
		new web3._extend.Method({
			name: 'startRPC',
			call: 'admin_startRPC',
			params: 4
		}),
		new web3._extend.Method({
			name: 'stopRPC',
			call: 'admin_stopRPC',
			params: 0
		}),
		new web3._extend.Method({
			name: 'startWS',
			call: 'admin_startWS',
			params: 4
		}),
		new web3._extend.Method({
			name: 'stopWS',
			call: 'admin_stopWS',
			params: 0
		}),
		new web3._extend.Method({
			name: 'setGlobalRegistrar',
			call: 'admin_setGlobalRegistrar',
			params: 2
		}),
		new web3._extend.Method({
			name: 'setHashReg',
			call: 'admin_setHashReg',
			params: 2
		}),
		new web3._extend.Method({
			name: 'setUrlHint',
			call: 'admin_setUrlHint',
			params: 2
		}),
		new web3._extend.Method({
			name: 'saveInfo',
			call: 'admin_saveInfo',
			params: 2
		}),
		new web3._extend.Method({
			name: 'register',
			call: 'admin_register',
			params: 3
		}),
		new web3._extend.Method({
			name: 'registerUrl',
			call: 'admin_registerUrl',
			params: 3
		}),
		new web3._extend.Method({
			name: 'startNatSpec',
			call: 'admin_startNatSpec',
			params: 0
		}),
		new web3._extend.Method({
			name: 'stopNatSpec',
			call: 'admin_stopNatSpec',
			params: 0
		}),
		new web3._extend.Method({
			name: 'getContractInfo',
			call: 'admin_getContractInfo',
			params: 1
		}),
		new web3._extend.Method({
			name: 'httpGet',
			call: 'admin_httpGet',
			params: 2
		})
	],
	properties:
	[
		new web3._extend.Property({
			name: 'nodeInfo',
			getter: 'admin_nodeInfo'
		}),
		new web3._extend.Property({
			name: 'peers',
			getter: 'admin_peers'
		}),
		new web3._extend.Property({
			name: 'datadir',
			getter: 'admin_datadir'
		})
	]
});
`

const Eth_JS = `
web3._extend({
	property: 'eth',
	methods:
	[
		new web3._extend.Method({
			name: 'sign',
			call: 'eth_sign',
			params: 2,
			inputFormatter: [web3._extend.utils.toAddress, null]
		}),
		new web3._extend.Method({
			name: 'resend',
			call: 'eth_resend',
			params: 3,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter, web3._extend.utils.fromDecimal, web3._extend.utils.fromDecimal]
		}),
		new web3._extend.Method({
			name: 'getNatSpec',
			call: 'eth_getNatSpec',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter]
		}),
		new web3._extend.Method({
			name: 'signTransaction',
			call: 'eth_signTransaction',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter]
		}),
		new web3._extend.Method({
			name: 'submitTransaction',
			call: 'eth_submitTransaction',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter]
		})
	],
	properties:
	[
		new web3._extend.Property({
			name: 'pendingTransactions',
			getter: 'eth_pendingTransactions'
		})
	]
});
`

const Net_JS = `
web3._extend({
	property: 'net',
	methods: [],
	properties:
	[
		new web3._extend.Property({
			name: 'version',
			getter: 'net_version'
		})
	]
});
`

const Debug_JS = `
web3._extend({
	property: 'debug',
	methods:
	[
		new web3._extend.Method({
			name: 'printBlock',
			call: 'debug_printBlock',
			params: 1
		}),
		new web3._extend.Method({
			name: 'getBlockRlp',
			call: 'debug_getBlockRlp',
			params: 1
		}),
		new web3._extend.Method({
			name: 'setHead',
			call: 'debug_setHead',
			params: 1
		}),
		new web3._extend.Method({
			name: 'traceBlock',
			call: 'debug_traceBlock',
			params: 2
		}),
		new web3._extend.Method({
			name: 'traceBlockByFile',
			call: 'debug_traceBlockByFile',
			params: 2
		}),
		new web3._extend.Method({
			name: 'traceBlockByNumber',
			call: 'debug_traceBlockByNumber',
			params: 2
		}),
		new web3._extend.Method({
			name: 'traceBlockByHash',
			call: 'debug_traceBlockByHash',
			params: 2
		}),
		new web3._extend.Method({
			name: 'seedHash',
			call: 'debug_seedHash',
			params: 1
		}),
		new web3._extend.Method({
			name: 'dumpBlock',
			call: 'debug_dumpBlock',
			params: 1
		}),
		new web3._extend.Method({
			name: 'metrics',
			call: 'debug_metrics',
			params: 1
		}),
		new web3._extend.Method({
			name: 'verbosity',
			call: 'debug_verbosity',
			params: 1
		}),
		new web3._extend.Method({
			name: 'vmodule',
			call: 'debug_vmodule',
			params: 1
		}),
		new web3._extend.Method({
			name: 'backtraceAt',
			call: 'debug_backtraceAt',
			params: 1,
		}),
		new web3._extend.Method({
			name: 'stacks',
			call: 'debug_stacks',
			params: 0,
			outputFormatter: console.log
		}),
		new web3._extend.Method({
			name: 'cpuProfile',
			call: 'debug_cpuProfile',
			params: 2
		}),
		new web3._extend.Method({
			name: 'startCPUProfile',
			call: 'debug_startCPUProfile',
			params: 1
		}),
		new web3._extend.Method({
			name: 'stopCPUProfile',
			call: 'debug_stopCPUProfile',
			params: 0
		}),
		new web3._extend.Method({
			name: 'trace',
			call: 'debug_trace',
			params: 2
		}),
		new web3._extend.Method({
			name: 'startTrace',
			call: 'debug_startTrace',
			params: 1
		}),
		new web3._extend.Method({
			name: 'stopTrace',
			call: 'debug_stopTrace',
			params: 0
		}),
		new web3._extend.Method({
			name: 'blockProfile',
			call: 'debug_blockProfile',
			params: 2
		}),
		new web3._extend.Method({
			name: 'setBlockProfileRate',
			call: 'debug_setBlockProfileRate',
			params: 1
		}),
		new web3._extend.Method({
			name: 'writeBlockProfile',
			call: 'debug_writeBlockProfile',
			params: 1
		}),
		new web3._extend.Method({
			name: 'writeMemProfile',
			call: 'debug_writeMemProfile',
			params: 1
		}),
		new web3._extend.Method({
			name: 'traceTransaction',
			call: 'debug_traceTransaction',
			params: 2
		})
	],
	properties: []
});
`

const Miner_JS = `
web3._extend({
	property: 'miner',
	methods:
	[
		new web3._extend.Method({
			name: 'start',
			call: 'miner_start',
			params: 1
		}),
		new web3._extend.Method({
			name: 'stop',
			call: 'miner_stop',
			params: 1
		}),
		new web3._extend.Method({
			name: 'setEtherbase',
			call: 'miner_setEtherbase',
			params: 1,
			inputFormatter: [web3._extend.formatters.formatInputInt],
			outputFormatter: web3._extend.formatters.formatOutputBool
		}),
		new web3._extend.Method({
			name: 'setExtra',
			call: 'miner_setExtra',
			params: 1
		}),
		new web3._extend.Method({
			name: 'setGasPrice',
			call: 'miner_setGasPrice',
			params: 1,
			inputFormatter: [web3._extend.utils.fromDecial]
		}),
		new web3._extend.Method({
			name: 'startAutoDAG',
			call: 'miner_startAutoDAG',
			params: 0
		}),
		new web3._extend.Method({
			name: 'stopAutoDAG',
			call: 'miner_stopAutoDAG',
			params: 0
		}),
		new web3._extend.Method({
			name: 'makeDAG',
			call: 'miner_makeDAG',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputDefaultBlockNumberFormatter]
		})
	],
	properties: []
});
`

const Shh_JS = `
web3._extend({
	property: 'shh',
	methods: [],
	properties:
	[
		new web3._extend.Property({
			name: 'version',
			getter: 'shh_version'
		})
	]
});
`
