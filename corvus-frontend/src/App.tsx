import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import Header from "./components/layout/Header";
import Footer from "./components/layout/Footer";
import LandingPage from "./pages/LandingPage";
import DeploymentViewerPage from "./pages/DeploymentViewerPage";
import { ToastProvider } from "./components/shared/Toast";

/** Root app component with routing and global providers */
function App() {
  return (
    <BrowserRouter>
      <ToastProvider>
        <div className="min-h-screen flex flex-col bg-white text-black">
          <Header />
          <Routes>
            <Route path="/" element={<LandingPage />} />
            <Route path="/d/:id" element={<DeploymentViewerPage />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
          <Footer />
        </div>
      </ToastProvider>
    </BrowserRouter>
  );
}

export default App;
