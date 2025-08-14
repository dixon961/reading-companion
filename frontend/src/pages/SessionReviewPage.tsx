import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getSession, exportSession, downloadFile, getSessionContent, getSessionMarkdown } from '../api/session';
import TwoPanelLayout from '../components/TwoPanelLayout';
import MarkdownRenderer from '../components/MarkdownRenderer';
import type { SessionData, SessionContent } from '../api/session';

// Component to render session content from JSON
const JSONRenderer: React.FC<{ content: SessionContent }> = ({ content }) => {
  return (
    <div className="json-renderer">
      <header className="json-header">
        <h1>{content.session.name}</h1>
        <p className="session-date">
          Дата разбора: {new Date(content.session.created_at).toLocaleDateString('ru-RU')}
        </p>
      </header>
      
      <div className="content-separator"></div>
      
      <div className="highlights-container">
        {content.highlights.map((highlight, index) => (
          <div key={index} className="highlight-block">
            <blockquote className="highlight-text">
              {highlight.text}
            </blockquote>
            
            <div className="interaction-container">
              <p className="question-text">
                <strong>Вопрос ассистента:</strong> {highlight.question}
              </p>
              
              {highlight.answered ? (
                <div className="answer-container">
                  <p className="answer-text">{highlight.answer}</p>
                </div>
              ) : (
                <p className="unanswered-text">Ответ не предоставлен</p>
              )}
            </div>
            
            {index < content.highlights.length - 1 && (
              <div className="content-separator"></div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};

const SessionReviewPage: React.FC = () => {
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();
  
  const [sessionData, setSessionData] = useState<SessionData | null>(null);
  const [sessionContent, setSessionContent] = useState<SessionContent | null>(null);
  const [sessionMarkdown, setSessionMarkdown] = useState<string>('');
  const [viewMode, setViewMode] = useState<'json' | 'markdown'>('markdown'); // Toggle between JSON and Markdown
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
        // First get basic session data
        const sessionInfo: SessionData = await getSession(sessionId);
        setSessionData(sessionInfo);
        
        // Then get the JSON content for review
        const content: SessionContent = await getSessionContent(sessionId);
        setSessionContent(content);
        
        // Finally get the markdown content for review
        const markdown: string = await getSessionMarkdown(sessionId);
        setSessionMarkdown(markdown);
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

  const toggleViewMode = () => {
    setViewMode(viewMode === 'json' ? 'markdown' : 'json');
  };

  if (isLoading) {
    return (
      <TwoPanelLayout>
        <div className="session-review-page">
          <div className="loading-container">
            <div className="spinner"></div>
            <p>Loading session data...</p>
          </div>
        </div>
      </TwoPanelLayout>
    );
  }

  if (error) {
    return (
      <TwoPanelLayout>
        <div className="session-review-page">
          <div className="error-container">
            <h2>Error</h2>
            <p>{error}</p>
            <button onClick={() => navigate('/')}>Back to Home</button>
          </div>
        </div>
      </TwoPanelLayout>
    );
  }

  if (!sessionData || !sessionContent) {
    return (
      <TwoPanelLayout>
        <div className="session-review-page">
          <div className="error-container">
            <h2>Error</h2>
            <p>No session data available</p>
            <button onClick={() => navigate('/')}>Back to Home</button>
          </div>
        </div>
      </TwoPanelLayout>
    );
  }

  return (
    <TwoPanelLayout>
      <div className="session-review-page">
        <header className="review-header">
          <h1>{sessionData.name}</h1>
          <p className="session-status">Completed Session</p>
        </header>
        
        <main className="review-main">
          <div className="review-controls">
            <button 
              onClick={toggleViewMode} 
              className="toggle-view-btn"
            >
              Switch to {viewMode === 'json' ? 'Markdown' : 'JSON'} View
            </button>
          </div>
          
          <div className="review-content">
            {viewMode === 'json' ? (
              <JSONRenderer content={sessionContent} />
            ) : (
              <MarkdownRenderer markdown={sessionMarkdown} />
            )}
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
    </TwoPanelLayout>
  );
};

export default SessionReviewPage;