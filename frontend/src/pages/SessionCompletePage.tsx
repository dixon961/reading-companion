import React, { useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getSession, exportSession, downloadFile } from '../api/session';
import TwoPanelLayout from '../components/TwoPanelLayout';
import type { SessionData } from '../api/session';
import { useLanguage } from '../i18n/LanguageContext';

const SessionCompletePage: React.FC = () => {
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();
  const { t } = useLanguage();
  const refreshSessionListRef = useRef<(() => void) | null>(null);
  
  const [sessionData, setSessionData] = React.useState<SessionData | null>(null);
  const [isLoading, setIsLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);

  React.useEffect(() => {
    const fetchSessionData = async () => {
      if (!sessionId) {
        setError(t('errors.generic'));
        setIsLoading(false);
        return;
      }
      
      try {
        const sessionInfo: SessionData = await getSession(sessionId);
        setSessionData(sessionInfo);
        // Refresh the session list to update the status
        if (refreshSessionListRef.current) {
          refreshSessionListRef.current();
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : t('errors.generic'));
      } finally {
        setIsLoading(false);
      }
    };

    fetchSessionData();
  }, [sessionId, t]);

  const handleDownloadSummary = async () => {
    if (!sessionId) return;
    
    try {
      const markdownContent = await exportSession(sessionId);
      const filename = `${sessionData?.name || 'session'}_export.md`;
      downloadFile(markdownContent, filename);
    } catch (err) {
      console.error(t('errors.generic'), err);
      alert(t('complete.downloadError'));
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
      <TwoPanelLayout onSessionListRefresh={() => { refreshSessionListRef.current = () => window.location.reload(); }}>
        <div className="session-complete-page">
          <div className="loading-container">
            <div className="spinner"></div>
            <p>{t('common.loading')}</p>
          </div>
        </div>
      </TwoPanelLayout>
    );
  }

  if (error) {
    return (
      <TwoPanelLayout onSessionListRefresh={() => { refreshSessionListRef.current = () => window.location.reload(); }}>
        <div className="session-complete-page">
          <div className="error-container">
            <h2>{t('common.error')}</h2>
            <p>{error}</p>
            <button onClick={() => navigate('/')} className="secondary-btn">{t('common.back')}</button>
          </div>
        </div>
      </TwoPanelLayout>
    );
  }

  if (!sessionData) {
    return (
      <TwoPanelLayout onSessionListRefresh={() => { refreshSessionListRef.current = () => window.location.reload(); }}>
        <div className="session-complete-page">
          <div className="error-container">
            <h2>{t('common.error')}</h2>
            <p>{t('session.noSessionData')}</p>
            <button onClick={() => navigate('/')} className="secondary-btn">{t('common.back')}</button>
          </div>
        </div>
      </TwoPanelLayout>
    );
  }

  return (
    <TwoPanelLayout onSessionListRefresh={() => { refreshSessionListRef.current = () => window.location.reload(); }}>
      <div className="session-complete-page">
        <header className="complete-header">
          <h1>{t('complete.sessionCompleted')}</h1>
          <p>{t('complete.congratulations').replace('{sessionData.name}', sessionData.name)}</p>
        </header>
        
        <main className="complete-main">
          <div className="success-icon">
            <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="#4CAF50" strokeWidth="2">
              <circle cx="12" cy="12" r="10"></circle>
              <path d="M8 12l2 2 4-4"></path>
            </svg>
          </div>
          
          <div className="session-stats">
            <p><strong>{sessionData.total_highlights}</strong> {t('complete.highlightsProcessed')}</p>
          </div>
          
          <div className="complete-actions">
            <button onClick={handleDownloadSummary} className="primary-btn">
              {t('complete.downloadSummary')}
            </button>
            <button onClick={handleStartNewSession} className="secondary-btn">
              {t('common.startNewSession')}
            </button>
            <button onClick={handleViewHistory} className="tertiary-btn">
              {t('complete.backToHistory')}
            </button>
          </div>
        </main>
      </div>
    </TwoPanelLayout>
  );
};

export default SessionCompletePage;