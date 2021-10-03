0. init typescript env.
#ref.
#https://nijialin.com/2020/09/19/how-to-build-typescript/

npm init
npm install typescript @types/node

#tsconfig.json
npx tsc --init

npm install ts-node nodemon --save-dev

#package.json
"scripts": {
    "build": "tsc",
    "start": "nodemon index.ts"
},

#jest testing
npm install -D jest ts-jest @types/jest

1. start up nats / nats jetstream
nats-server.exe -c js.conf

#official doc.
#https://docs.nats.io/jetstream/administration

#ref url
#https://www.jianshu.com/p/27a49b9d4306

#jetstream sample code
#https://github.com/nats-io/nats.deno/blob/main/jetstream.md

#official docker hub
#https://hub.docker.com/_/nats

