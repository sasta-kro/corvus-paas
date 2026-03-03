import corvusLogo from '/corvus.svg'
import './App.css'

function App() {
  const message = document.getElementById('root')?.getAttribute('data-message') || 'Hello, World!'

  return (
      <div className="page-container">
          <div className="message-container">
              <a href="https://github.com/sasta-kro/corvus-paas" target="_blank" rel="noopener noreferrer">
                  <img src={corvusLogo} className="corvus-logo" alt="Corvus logo"/>
              </a>
              <p className="message-text">{message}</p>
              <p className="message-flavor">
                  This page was created and deployed on{' '}
                  <a href="https://github.com/sasta-kro/corvus-paas" target="_blank"
                     rel="noopener noreferrer">Corvus</a>,
                  a self-hosted platform that goes from source to live URL in seconds.
              </p>
              <p className="message-flavor" >The website link workss everywhere, share it with your friends!</p>
              <p className="message-credits">
                  Built by <a href="https://github.com/sasta-kro" target="_blank"
                              rel="noopener noreferrer">sasta-kro</a>
              </p>
          </div>
      </div>
  )
}

export default App
