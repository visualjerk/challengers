<script setup lang="ts">
import { gameClient, GameEvent } from './game-api'
import { onMounted, onBeforeUnmount, ref, unref, computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const gameId = computed(() => route.params.id as string)

const joined = ref(false)
async function join() {
  const { response } = await gameClient.playerAction({
    gameId: unref(gameId),
    message: {
      oneofKind: 'playerJoin',
      playerJoin: {},
    },
  })
  if (response.response.oneofKind === 'error') {
    console.error('error joining:', response.response.error.message)
    return
  }
  joined.value = true
}
async function leave() {
  const { response } = await gameClient.playerAction({
    gameId: unref(gameId),
    message: {
      oneofKind: 'playerLeave',
      playerLeave: {},
    },
  })
  if (response.response.oneofKind === 'error') {
    console.error('error joining:', response.response.error.message)
    return
  }
  joined.value = false
}

const events = ref<GameEvent[]>([])
let unsubscribe: any

onMounted(() => {
  const gameEvents = gameClient.gameEvents({
    gameId: unref(gameId),
  })
  unsubscribe = gameEvents.responses.onMessage((event) => {
    events.value.unshift(event)
  })
})

onBeforeUnmount(() => unsubscribe())
</script>

<template>
  <div class="grid gap-2 p-3">
    <h1 class="text-xl">Game</h1>
    <div v-if="!joined">
      <button @click="join" class="bg-indigo-600 text-white p-2">
        Join Game
      </button>
    </div>
    <div v-else>
      <button @click="leave" class="bg-indigo-600 text-white p-2">
        Leave Game
      </button>
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
