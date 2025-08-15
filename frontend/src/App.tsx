import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import HomePage from './pages/HomePage';
import SessionPage from './pages/SessionPage';
import SessionReviewPage from './pages/SessionReviewPage';
import SessionCompletePage from './pages/SessionCompletePage';
import './App.css';

function App() {
  return (
    <Router>
      <div className="app">
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/session/:sessionId" element={<SessionPage />} />
          <Route path="/review/:sessionId" element={<SessionReviewPage />} />
          <Route path="/complete/:sessionId" element={<SessionCompletePage />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;
