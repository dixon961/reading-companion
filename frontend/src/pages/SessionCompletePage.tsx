import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getSession, exportSession, downloadFile } from '../api/session';
import type { SessionData } from '../api/session';

const SessionCompletePage: React.FC = () => {
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();
  
  const [sessionData, setSessionData] = React.useState<SessionData | null>(null);
  const [isLoading, setIsLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);

  React.useEffect(() => {
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

  const handleDownloadSummary = async () => {
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

  const handleStartNewSession = () => {
    navigate('/');
  };

  const handleViewHistory = () => {
    navigate('/');
  };

  if (isLoading) {
    return (
      <div className="session-complete-page">
        <div className="loading-container">
          <div className="spinner"></div>
          <p>Loading session data...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="session-complete-page">
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
      <div className="session-complete-page">
        <div className="error-container">
          <h2>Error</h2>
          <p>No session data available</p>
          <button onClick={() => navigate('/')}>Back to Home</button>
        </div>
      </div>
    );
  }

  return (
    <div className="session-complete-page">
      <header className="complete-header">
        <h1>Сессия успешно завершена!</h1>
        <p>Поздравляем! Вы успешно обработали все пометки из книги "{sessionData.name}".</p>
      </header>
      
      <main className="complete-main">
        <div className="success-icon">
          <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="#4CAF50" strokeWidth="2">
            <circle cx="12" cy="12" r="10"></circle>
            <path d="M8 12l2 2 4-4"></path>
          </svg>
        </div>
        
        <div className="session-stats">
          <p><strong>{sessionData.total_highlights}</strong> пометок обработано</p>
        </div>
        
        <div className="complete-actions">
          <button onClick={handleDownloadSummary} className="primary-btn">
            Скачать конспект (.md)
          </button>
          <button onClick={handleStartNewSession} className="secondary-btn">
            Начать новую сессию
          </button>
          <button onClick={handleViewHistory} className="tertiary-btn">
            Вернуться к истории
          </button>
        </div>
      </main>
    </div>
  );
};

export default SessionCompletePage;