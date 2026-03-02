(function(){const o=document.createElement("link").relList;if(o&&o.supports&&o.supports("modulepreload"))return;for(const e of document.querySelectorAll('link[rel="modulepreload"]'))s(e);new MutationObserver(e=>{for(const t of e)if(t.type==="childList")for(const a of t.addedNodes)a.tagName==="LINK"&&a.rel==="modulepreload"&&s(a)}).observe(document,{childList:!0,subtree:!0});function r(e){const t={};return e.integrity&&(t.integrity=e.integrity),e.referrerPolicy&&(t.referrerPolicy=e.referrerPolicy),e.crossOrigin==="use-credentials"?t.credentials="include":e.crossOrigin==="anonymous"?t.credentials="omit":t.credentials="same-origin",t}function s(e){if(e.ep)return;e.ep=!0;const t=r(e);fetch(e.href,t)}})();const i="/vite.svg",n="/corvus.svg";function u(c){let o=0;const r=s=>{o=s,c.innerHTML=`count is ${o}`};c.addEventListener("click",()=>r(o+1)),r(0)}document.querySelector("#app").innerHTML=`
  <div>
    <div class="logos">
      <a href="https://vite.dev" target="_blank">
        <img src="${i}" class="logo" alt="Vite logo" />
      </a>
      <a href="https://github.com/sasta-kro/corvus-paas" target="_blank">
        <img src="${n}" class="logo corvus" alt="Corvus logo" />
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
`;u(document.querySelector("#counter"));
