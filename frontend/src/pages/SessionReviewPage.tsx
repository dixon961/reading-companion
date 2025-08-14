import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getSession, exportSession, downloadFile } from '../api/session';
import type { SessionData } from '../api/session';

const SessionReviewPage: React.FC = () => {
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();
  
  const [sessionData, setSessionData] = useState<SessionData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchSessionData = async () => {
      if (!sessionId) {
        setError('No session ID provided');
        setIsLoading(false);
        return;
      }
      
      try {
        const sessionInfo: SessionData = await getSession(sessionId);
        setSessionData(sessionInfo);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load session data');
      } finally {
        setIsLoading(false);
      }
    };

    fetchSessionData();
  }, [sessionId]);

  const handleDownload = async () => {
    if (!sessionId) return;
    
    try {
      const markdownContent = await exportSession(sessionId);
      const filename = `${sessionData?.name || 'session'}_export.md`;
      downloadFile(markdownContent, filename);
    } catch (err) {
      console.error('Failed to download summary:', err);
      alert('Failed to download summary. Please try again.');
    }
  };

  const handleNewSession = () => {
    navigate('/');
  };

  if (isLoading) {
    return (
      <div className="session-review-page">
        <div className="loading-container">
          <div className="spinner"></div>
          <p>Loading session data...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="session-review-page">
        <div className="error-container">
          <h2>Error</h2>
          <p>{error}</p>
          <button onClick={() => navigate('/')}>Back to Home</button>
        </div>
      </div>
    );
  }

  if (!sessionData) {
    return (
      <div className="session-review-page">
        <div className="error-container">
          <h2>Error</h2>
          <p>No session data available</p>
          <button onClick={() => navigate('/')}>Back to Home</button>
        </div>
      </div>
    );
  }

  return (
    <div className="session-review-page">
      <header className="review-header">
        <h1>{sessionData.name}</h1>
        <p className="session-status">Completed Session</p>
      </header>
      
      <main className="review-main">
        <div className="review-content">
          <h2>Session Summary</h2>
          <p>Total highlights processed: {sessionData.total_highlights}</p>
          
          {/* TODO: Add actual session review content here */}
          <div className="review-placeholder">
            <p>Detailed session review content will be displayed here.</p>
            <p>This will include all the questions and answers from your session.</p>
          </div>
        </div>
        
        <div className="review-actions">
          <button onClick={handleDownload} className="download-btn">
            Download Summary (.md)
          </button>
          <button onClick={handleNewSession}>
            Start New Session
          </button>
        </div>
      </main>
    </div>
  );
};

export default SessionReviewPage;