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
  if (response.response.oneofKind === 'error') {
    console.error('error joining:', response.response.error.message)
    return
  }
  name.value = ''
}

const events = ref<GameEvent[]>([])
const gameEvents = gameClient.gameEvents({})
const unsubscribe = gameEvents.responses.onMessage((event) => {
  events.value.push(event)
})
onBeforeUnmount(() => unsubscribe())
</script>

<template>
  <div class="grid gap-2 p-3">
    <div>
      <h2>Join Game</h2>
      <form @submit.prevent="join">
        <input
          v-model="name"
          placeholder="How shall we call you?"
          class="p-2 border border-slate-400"
        />
      </form>
    </div>
    <div class="events">
      <h2>Events</h2>
      <div v-for="(event, index) in events" :key="index">
        <h3>{{ event.date }}</h3>
        <pre>{{ event.message }}</pre>
      </div>
    </div>
  </div>
</template>
