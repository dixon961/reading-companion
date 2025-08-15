import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import SessionListItem from './SessionListItem';
import LanguageSelector from './LanguageSelector';
import { listSessions, updateSessionName, deleteSession } from '../api/session';
import type { SessionMetadata } from '../api/session';
import { useLanguage } from '../i18n/LanguageContext';

interface TwoPanelLayoutProps {
  children: React.ReactNode;
  showSessionList?: boolean;
  onSessionListRefresh?: () => void;
}

const TwoPanelLayout: React.FC<TwoPanelLayoutProps> = ({ 
  children, 
  showSessionList = true,
  onSessionListRefresh
}) => {
  const [sessions, setSessions] = useState<SessionMetadata[]>([]);
  const [sessionsLoading, setSessionsLoading] = useState(true);
  const [sessionsError, setSessionsError] = useState<string | null>(null);
  const navigate = useNavigate();
  const { t } = useLanguage();

  // Fetch sessions when component mounts
  useEffect(() => {
    if (showSessionList) {
      fetchSessions();
    }
  }, [showSessionList]);

  const fetchSessions = async () => {
    setSessionsLoading(true);
    setSessionsError(null);
    
    try {
      const sessionList = await listSessions();
      setSessions(sessionList);
    } catch (err) {
      setSessionsError(err instanceof Error ? err.message : 'Failed to load sessions');
    } finally {
      setSessionsLoading(false);
    }
  };

  const handleRenameSession = async (sessionId: string, newName: string) => {
    try {
      await updateSessionName(sessionId, { name: newName });
      // Refresh the session list
      await fetchSessions();
      // Notify parent component if needed
      if (onSessionListRefresh) {
        onSessionListRefresh();
      }
    } catch (err) {
      throw err;
    }
  };

  const handleDeleteSession = async (sessionId: string) => {
    try {
      await deleteSession(sessionId);
      // Refresh the session list
      await fetchSessions();
      // Notify parent component if needed
      if (onSessionListRefresh) {
        onSessionListRefresh();
      }
    } catch (err) {
      throw err;
    }
  };

  const handleNewSession = () => {
    navigate('/');
  };

  return (
    <div className="two-panel-layout">
      {showSessionList && (
        <aside className="session-sidebar">
          <div className="sidebar-header">
            <h2>{t('common.sessionHistory')}</h2>
            <div className="sidebar-controls">
              <button className="new-session-btn" onClick={handleNewSession}>
                {t('common.startNewSession')}
              </button>
              <LanguageSelector className="sidebar-language-selector" />
            </div>
          </div>
          
          <div className="session-list-container">
            {sessionsLoading ? (
              <div className="loading-container">
                <div className="spinner"></div>
                <p>{t('common.loading')}</p>
              </div>
            ) : sessionsError ? (
              <div className="error-container">
                <p>{t('common.error')} loading sessions: {sessionsError}</p>
                <button onClick={fetchSessions}>{t('common.retry')}</button>
              </div>
            ) : sessions.length > 0 ? (
              <div className="session-list">
                {sessions.map((session) => (
                  <SessionListItem 
                    key={session.id} 
                    session={session} 
                    onRename={handleRenameSession}
                    onDelete={handleDeleteSession}
                  />
                ))}
              </div>
            ) : (
              <div className="empty-state">
                <p>{t('common.noSessionsMessage')}</p>
              </div>
            )}
          </div>
        </aside>
      )}
      
      <main className="main-content">
        {children}
      </main>
    </div>
  );
};

export default TwoPanelLayout;