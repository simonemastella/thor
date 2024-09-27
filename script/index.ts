import { Transaction, secp256k1 } from "thor-devkit";

const clauses = [
    {
        to: "0x7567d83b7b8d80addcb281a71d54fc7b3364ffed",
        value: 10000,
        data: "0x",
    },
];
// Transaction.Body
let body: any = {
    chainTag: "0xf5",
    blockRef: "0x0000000000000000",
    expiration: 32,
    clauses: clauses,
    gasPriceCoef: 128,
    gas: Transaction.intrinsicGas(clauses),
    dependsOn: null,
    nonce: 12345678,
};

const tx = new Transaction(body);
const signingHash = tx.signingHash();
tx.signature = secp256k1.sign(
    signingHash,
    Buffer.from(
        "99f0500549792796c14fed62011a51081dc5b5e68fe8bd8a13b86be829c4fd36",
        "hex"
    )
);

const raw = tx.encode();
const decoded = Transaction.decode(raw);
const response = await fetch("http://localhost:8669/transactions/schedule", {
    method: "POST",
    body: JSON.stringify({
        raw: "0x" + raw.toString("hex"),
        time: "2012-01-05T15:04:05Z",
    }),
    headers: { "Content-Type": "application/json" },
});
console.log("decoded", decoded);
console.log("response", response);
