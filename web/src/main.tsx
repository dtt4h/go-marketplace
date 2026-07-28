import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'

import App from './app/App.tsx'
import { configureAuthInterceptor } from './features/auth/api/configureAuthInterceptor'
import './shared/styles/reset.scss'
import './shared/styles/global.scss'

configureAuthInterceptor()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </StrictMode>,
)
