<script setup lang="ts">
import { GameClient } from '../../api/game.client'
import { GrpcWebFetchTransport } from '@protobuf-ts/grpcweb-transport'
import { useRouter } from 'vue-router'

const rpcTransport = new GrpcWebFetchTransport({
  baseUrl: 'http://0.0.0.0:50051',
})
const gameClient = new GameClient(rpcTransport)
const router = useRouter()

async function createGame() {
  const { response } = await gameClient.createGame({})
  router.push(`/game/${response.id}`)
}
</script>

<template>
  <div class="grid gap-2 p-3">
    <h1 class="text-xl">Home</h1>
    <div>
      <button @click="createGame">Create New Game</button>
    </div>
  </div>
</template>
