<script setup lang="ts">
import { onBeforeUnmount, ref, unref } from 'vue'
import { GameEvent } from '../api/game'
import { GameClient } from '../api/game.client'
import { GrpcWebFetchTransport } from '@protobuf-ts/grpcweb-transport'

const rpcTransport = new GrpcWebFetchTransport({
  baseUrl: 'http://0.0.0.0:50051',
})
const gameClient = new GameClient(rpcTransport)

const name = ref('')
async function join() {
  const { response } = await gameClient.playerAction({
    message: {
      oneofKind: 'playerJoined',
      playerJoined: {
        id: 'hans',
        name: unref(name),
      },
    },
  })
  console.log('join response', response)
}

const events = ref<GameEvent[]>([])
const gameEvents = gameClient.gameEvents({})
const unsubscribe = gameEvents.responses.onMessage((event) => {
  events.value.push(event)
})
onBeforeUnmount(() => unsubscribe())
</script>

<template>
  <div>
    <div class="game">
      <form @submit.prevent="join">
        <input v-model="name" placeholder="How shall we call you?" />
      </form>
      <header>
        <h2>Events</h2>
      </header>
      <div class="events">
        <div v-for="(event, index) in events" :key="index">
          <h3>{{ event.date }}</h3>
          <pre>{{ event.message }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.game {
  display: grid;
  grid-template-rows: auto 1fr auto;
  height: 100vh;
  width: 100%;
}

.events {
  overflow-y: auto;
}

form {
  padding: 1rem;
}

input {
  padding: 1rem;
  font-size: 1.2rem;
  width: 100%;
}
</style>
