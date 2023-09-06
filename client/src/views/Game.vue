<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, unref, computed } from 'vue'
import { GameEvent } from '../../api/game'
import { gameClient } from '../api/game'
import { useRoute } from 'vue-router'

const route = useRoute()
const gameId = computed(() => route.params.id as string)

const name = ref('')
async function join() {
  const { response } = await gameClient.playerAction({
    gameId: unref(gameId),
    message: {
      oneofKind: 'playerJoin',
      playerJoin: {
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
    <div>
      <h2>Join This Game</h2>
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
