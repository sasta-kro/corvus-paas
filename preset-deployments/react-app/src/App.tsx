import { useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import corvusLogo from '/corvus.svg'
import './App.css'

function App() {
  const [count, setCount] = useState(0)

  return (
    <>
      <div className="logos">
        <a href="https://vite.dev" target="_blank">
          <img src={viteLogo} className="logo" alt="Vite logo" />
        </a>
        <a href="https://react.dev" target="_blank">
          <img src={reactLogo} className="logo react" alt="React logo" />
        </a>
        <a href="https://github.com/sasta-kro/corvus-paas" target="_blank" rel="noopener noreferrer">
          <img src={corvusLogo} className="logo corvus" alt="Corvus logo" />
        </a>
      </div>
      <h1>Vite + React + Corvus</h1>
      <div className="card">
        <button onClick={() => setCount((count) => count + 1)}>
          count is {count}
        </button>
      </div>
      <p className="flavor">
        Deployed on <a href="https://github.com/sasta-kro/corvus-paas" target="_blank" rel="noopener noreferrer">Corvus</a>,
        a platform built from scratch in Go that orchestrates containers through the Docker SDK
        and routes traffic with Traefik. From push to live URL in seconds.
      </p>
      <p className="credits">
        Built by <a href="https://github.com/sasta-kro" target="_blank" rel="noopener noreferrer">sasta-kro</a>
      </p>
    </>
  )
}

export default App
