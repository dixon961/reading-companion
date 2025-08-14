import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getSession, exportSession, downloadFile, getSessionContent, getSessionMarkdown } from '../api/session';
import TwoPanelLayout from '../components/TwoPanelLayout';
import MarkdownRenderer from '../components/MarkdownRenderer';
import ReadOnlyHighlights from '../components/ReadOnlyHighlights';
import type { SessionData, SessionContent } from '../api/session';
import { useLanguage } from '../i18n/LanguageContext';

// Component to render session content from JSON
const JSONRenderer: React.FC<{ content: SessionContent; language: string }> = ({ content, language }) => {
  return (
    <div className="json-renderer">
      <header className="json-header">
        <h1>{content.session.name}</h1>
        <p className="session-date">
          {language === 'ru' ? 'Дата разбора: ' : 'Review Date: '}
          {new Date(content.session.created_at).toLocaleDateString(language === 'ru' ? 'ru-RU' : 'en-US')}
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
                <strong>{language === 'ru' ? 'Вопрос ассистента:' : 'Assistant Question:'} </strong> 
                {highlight.question}
              </p>
              
              {highlight.answered ? (
                <div className="answer-container">
                  <p className="answer-text">{highlight.answer}</p>
                </div>
              ) : (
                <p className="unanswered-text">
                  {language === 'ru' ? 'Ответ не предоставлен' : 'No answer provided'}
                </p>
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
  const { t, language } = useLanguage();
  const refreshSessionListRef = useRef<(() => void) | null>(null);
  
  const [sessionData, setSessionData] = useState<SessionData | null>(null);
  const [sessionContent, setSessionContent] = useState<SessionContent | null>(null);
  const [sessionMarkdown, setSessionMarkdown] = useState<string>('');
  const [viewMode, setViewMode] = useState<'highlights' | 'json' | 'markdown'>('highlights'); // Toggle between Highlights, JSON and Markdown
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchSessionData = async () => {
      if (!sessionId) {
        setError(t('errors.generic'));
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

  const handleDownload = async () => {
    if (!sessionId) return;
    
    try {
      const markdownContent = await exportSession(sessionId);
      const filename = `${sessionData?.name || 'session'}_export.md`;
      downloadFile(markdownContent, filename);
    } catch (err) {
      console.error(t('errors.generic'), err);
      alert(t('review.downloadError'));
    }
  };

  const handleNewSession = () => {
    navigate('/');
  };

  const getViewModeText = () => {
    switch (viewMode) {
      case 'highlights':
        return t('review.switchToJSON');
      case 'json':
        return t('review.switchToMarkdown');
      case 'markdown':
        return t('review.switchToHighlights');
      default:
        return t('review.switchToJSON');
    }
  };

  const toggleViewMode = () => {
    setViewMode(prevMode => {
      switch (prevMode) {
        case 'highlights':
          return 'json';
        case 'json':
          return 'markdown';
        case 'markdown':
          return 'highlights';
        default:
          return 'highlights';
      }
    });
  };

  if (isLoading) {
    return (
      <TwoPanelLayout onSessionListRefresh={() => { refreshSessionListRef.current = () => window.location.reload(); }}>
        <div className="session-review-page">
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
        <div className="session-review-page">
          <div className="error-container">
            <h2>{t('common.error')}</h2>
            <p>{error}</p>
            <button onClick={() => navigate('/')} className="secondary-btn">{t('common.back')}</button>
          </div>
        </div>
      </TwoPanelLayout>
    );
  }

  if (!sessionData || !sessionContent) {
    return (
      <TwoPanelLayout onSessionListRefresh={() => { refreshSessionListRef.current = () => window.location.reload(); }}>
        <div className="session-review-page">
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
      <div className="session-review-page">
        <header className="review-header">
          <h1>{sessionData.name}</h1>
          <p className="session-status">{t('common.completed')}</p>
        </header>
        
        <main className="review-main">
          <div className="review-controls">
            <button 
              onClick={toggleViewMode} 
              className="toggle-view-btn"
            >
              {getViewModeText()}
            </button>
          </div>
          
          <div className="review-content">
            {viewMode === 'highlights' ? (
              <ReadOnlyHighlights content={sessionContent} />
            ) : viewMode === 'json' ? (
              <JSONRenderer content={sessionContent} language={language} />
            ) : (
              <MarkdownRenderer markdown={sessionMarkdown} />
            )}
          </div>
          
          <div className="review-actions">
            <button onClick={handleDownload} className="download-btn">
              {t('review.downloadSummary')}
            </button>
            <button onClick={handleNewSession} className="secondary-btn">
              {t('common.startNewSession')}
            </button>
          </div>
        </main>
      </div>
    </TwoPanelLayout>
  );
};

export default SessionReviewPage;