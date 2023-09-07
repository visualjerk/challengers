import { GameClient } from '../../api/game.client'
import { GrpcWebFetchTransport } from '@protobuf-ts/grpcweb-transport'
import { API_BASE_URL } from '../config/api-config'
import { AuthenticationInterceptor } from '../account/account-api'

const rpcTransport = new GrpcWebFetchTransport({
  baseUrl: API_BASE_URL,
  interceptors: [AuthenticationInterceptor],
})
export const gameClient = new GameClient(rpcTransport)
