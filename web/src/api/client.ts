import axios from 'axios'

export const apiClient = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-type': 'application/json',
  },
  timeout: 10_000,
})
