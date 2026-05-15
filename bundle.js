import { SigningStargateClient, defaultRegistryTypes, AminoTypes } from "@cosmjs/stargate";
import { Registry, encodePubkey } from "@cosmjs/proto-signing";
import { TxRaw } from "cosmjs-types/cosmos/tx/v1beta1/tx";
import { makeSignDoc, encodeSecp256k1Pubkey } from "@cosmjs/amino";

const MsgCreateTaskTypeUrl = "/portalchain.tasks.MsgCreateTask";

function encodeVarint(value) {
  const bytes = [];
  while (value > 0x7f) {
    bytes.push((value & 0x7f) | 0x80);
    value >>>= 7;
  }
  bytes.push(value & 0x7f);
  return new Uint8Array(bytes);
}
function encodeField(fieldNum, value) {
  if (!value) return new Uint8Array(0);
  const strBytes = new TextEncoder().encode(value);
  const tag = new Uint8Array([(fieldNum << 3) | 2]);
  const len = encodeVarint(strBytes.length);
  const result = new Uint8Array(tag.length + len.length + strBytes.length);
  result.set(tag, 0);
  result.set(len, tag.length);
  result.set(strBytes, tag.length + len.length);
  return result;
}
function encodeMsgCreateTaskBytes({ creator, queryHash, queryUrl, taskType }) {
  const parts = [
    encodeField(1, creator),
    encodeField(2, queryHash),
    encodeField(3, queryUrl || ''),
    encodeField(4, taskType),
  ];
  const total = parts.reduce((a, b) => a + b.length, 0);
  const result = new Uint8Array(total);
  let offset = 0;
  for (const part of parts) { result.set(part, offset); offset += part.length; }
  return result;
}

const MsgCreateTaskCodec = {
  encode: (msg) => {
    console.log('[CosmJS] MsgCreateTaskCodec.encode called with:', JSON.stringify(msg));
    const bytes = encodeMsgCreateTaskBytes({
      creator: msg.creator,
      queryHash: msg.queryHash || msg.query_hash || '',
      queryUrl: msg.queryUrl || msg.query_url || '',
      taskType: msg.taskType || msg.task_type || '',
    });
    console.log('[CosmJS] encoded bytes length:', bytes.length);
    return { finish: () => bytes };
  },
  decode: () => ({}),
  fromPartial: (msg) => msg,
};

window.broadcastWithCosmJS = async function ({ rpcUrl, walletAddress, offlineSigner, queryHash, queryUrl, taskType, accountNumber: extAccountNumber, sequence: extSequence }) {
  const accountNumber = extAccountNumber ? parseInt(extAccountNumber) : 0;
  const sequence = extSequence ? parseInt(extSequence) : 0;
  console.log('[CosmJS] accountNumber:', accountNumber, 'sequence:', sequence);

  const registry = new Registry([
    ...defaultRegistryTypes,
    [MsgCreateTaskTypeUrl, MsgCreateTaskCodec],
  ]);

  const aminoTypes = new AminoTypes({
    "/portalchain.tasks.MsgCreateTask": {
      aminoType: "tasks/MsgCreateTask",
      toAmino: ({ creator, queryHash, queryUrl, taskType }) => ({ creator, query_hash: queryHash, query_url: queryUrl, task_type: taskType }),
      fromAmino: ({ creator, query_hash, query_url, task_type }) => ({ creator, queryHash: query_hash, queryUrl: query_url, taskType: task_type }),
    },
  });

  // Override getSequence to return our fresh values
  const signerWithFreshSequence = {
    ...offlineSigner,
    getAccounts: () => offlineSigner.getAccounts(),
    signAmino: async (signerAddress, signDoc) => {
      // Override accountNumber and sequence in signDoc
      const fixedSignDoc = {
        ...signDoc,
        account_number: String(accountNumber),
        sequence: String(sequence),
      };
      console.log('[CosmJS] fixed signDoc:', JSON.stringify(fixedSignDoc));
      const sigResult = await offlineSigner.signAmino(signerAddress, fixedSignDoc);
      console.log('[CosmJS] signAmino result signed:', JSON.stringify(sigResult.signed));
      if (sigResult.signed.account_number !== fixedSignDoc.account_number || sigResult.signed.sequence !== fixedSignDoc.sequence) {
        console.error('[CosmJS] MISMATCH! sent:', fixedSignDoc.account_number, fixedSignDoc.sequence, 'got back:', sigResult.signed.account_number, sigResult.signed.sequence);
      }
      return sigResult;
    },
  };

  const client = await SigningStargateClient.connectWithSigner(rpcUrl, offlineSigner, {
    registry,
    aminoTypes,
  });

  // Patch getSequence to return fresh values
  client.getSequence = async (addr) => { console.log("[CosmJS] getSequence called for", addr, "returning", accountNumber, sequence); return { accountNumber, sequence }; };

  const msgs = [{ typeUrl: MsgCreateTaskTypeUrl, value: { creator: walletAddress, queryHash, queryUrl: queryUrl || "", taskType } }];
  const fee = { amount: [{ denom: "udaai", amount: "5000" }], gas: "200000" };

  // Intercept to log tx bytes
  const origBroadcast = client.broadcastTx.bind(client);
  client.broadcastTx = async (txBytes) => {
    const txBase64 = btoa(Array.from(txBytes).map(b => String.fromCharCode(b)).join(''));
    console.log('[CosmJS] txBytes base64:', txBase64);
    // Decode and check via REST
    const checkResp = await fetch('https://api.portalchain.org/cosmos/tx/v1beta1/decode', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({tx_bytes: txBase64})
    });
    const checkJson = await checkResp.json();
    console.log('[CosmJS] decoded tx:', JSON.stringify(checkJson));
    try { const r = await origBroadcast(txBytes); console.log("[CosmJS] origBroadcast result:", JSON.stringify(r, (k,v) => typeof v === "bigint" ? v.toString() : v)); return r; } catch(e) { console.error("[CosmJS] origBroadcast error:", e.message); throw e; }
  };
  console.log('[CosmJS] calling sign with explicit signerData...');
  const memo = "DAAI Web UI pay-ask " + Date.now();
  const signerData = { accountNumber, sequence, chainId: "portalchain" };
  const signed = await client.sign(walletAddress, msgs, fee, memo, signerData);
  console.log('[CosmJS] signed ok, broadcasting...');
  try {
    const txBytes = TxRaw.encode(signed).finish();
    console.log('[CosmJS] txBytes length:', txBytes.length);
    const broadcastResult = await client.broadcastTx(Uint8Array.from(txBytes));
    console.log('[CosmJS] broadcast result:', JSON.stringify(broadcastResult, (k,v) => typeof v === 'bigint' ? v.toString() : v));
    return broadcastResult;
  } catch(e) {
    console.error('[CosmJS] broadcast error:', e.message, e.stack);
    throw e;
  }
};

console.log('[CosmJS bridge] ready');
