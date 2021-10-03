import { AckPolicy, connect, Empty, PubAck, StringCodec } from "nats";

const jetstreamClient = async () => {

const nc = await connect({ servers: "127.0.0.1:4222" });

const jsm = await nc.jetstreamManager();
// add a stream
//await jsm.streams.add({ name: "a", subjects: ["a.*"] });

var order = {
    OrderId : "testOrder",
    CustomerId : "c1",
    Status : "s1"
}

var str = JSON.stringify(order, null, 0);

var ret = new Uint8Array(str.length);
	for (var i = 0; i < str.length; i++) {
		ret[i] = str.charCodeAt(i);
	}


// create a jetstream client:
const js = nc.jetstream();

// to publish messages to a stream:
let pa = await js.publish("ORDER.created",ret);
//let pa = await js.publish("a.b");


// the jetstream returns an acknowledgement with the
// stream that captured the message, it's assigned sequence
// and whether the message is a duplicate.
const stream = pa.stream;
const seq = pa.seq;
const duplicate = pa.duplicate;

const psub = await js.pullSubscribe("ORDER.created", { config: { durable_name: "c" } });

let msg = await js.pull("ORDER", "c");
console.log(msg);
msg.ack();

let test = msg.data;

console.log(test.toString());
/*
const done = (async () => {
  for await (const m of psub) {
    console.log(`${m.info.stream}[${m.seq}]`);
    m.ack();
  }
})();
*/
/*
// More interesting is the ability to prevent duplicates
// on messages that are stored in the server. If
// you assign a message ID, the server will keep looking
// for the same ID for a configured amount of time, and
// reject messages that sport the same ID:
await js.publish("a.b", Empty, { msgID: "a" });

// you can also specify constraints that should be satisfied.
// For example, you can request the message to have as its
// last sequence before accepting the new message:
await js.publish("a.b", Empty, { expect: { lastMsgID: "a" } });
await js.publish("a.b", Empty, { expect: { lastSequence: 3 } });
// save the last sequence for this publish
pa = await js.publish("a.b", Empty, { expect: { streamName: "a" } });
// you can also mix the above combinations

// this stream here accepts wildcards, you can assert that the
// last message sequence recorded on a particular subject matches:
const buf: Promise<PubAck>[] = [];
for (let i = 0; i < 100; i++) {
  buf.push(js.publish("a.a", Empty));
}
await Promise.all(buf);
// if additional "a.b" has been recorded, this will fail
await js.publish("a.b", Empty, { expect: { lastSubjectSequence: pa.seq } });
*/
}

jetstreamClient()