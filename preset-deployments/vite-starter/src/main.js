import './style.css'
import viteLogo from '/vite.svg'
import corvusLogo from '/corvus.svg'
import { setupCounter } from './counter.js'

document.querySelector('#app').innerHTML = `
  <div>
    <div class="logos">
      <a href="https://vite.dev" target="_blank">
        <img src="${viteLogo}" class="logo" alt="Vite logo" />
      </a>
      <a href="https://github.com/sasta-kro/corvus-paas" target="_blank">
        <img src="${corvusLogo}" class="logo corvus" alt="Corvus logo" />
      </a>
    </div>
    <h1>Vite + Corvus</h1>
    <div class="card">
      <button id="counter" type="button"></button>
    </div>
    <p class="flavor">
      This site was deployed in seconds on
      <a href="https://github.com/sasta-kro/corvus-paas" target="_blank">Corvus</a>,
      a self-hosted PaaS that talks directly to the Docker daemon.
      No managed services, no abstraction layers.
    </p>
    <p class="credits">
      Built by <a href="https://github.com/sasta-kro" target="_blank">sasta-kro</a>
    </p>
  </div>
`

setupCounter(document.querySelector('#counter'))
