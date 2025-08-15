import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { processAnswer, getSession, regenerateQuestion } from '../api/session';
import TwoPanelLayout from '../components/TwoPanelLayout';
import type { CreateSessionResponse, ProcessAnswerRequest, ProcessAnswerResponse, SessionData, RegenerateQuestionRequest } from '../api/session';
import { useLanguage } from '../i18n/LanguageContext';

const SessionPage: React.FC = () => {
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();
  const { t } = useLanguage();
  const refreshSessionListRef = useRef<(() => void) | null>(null);
  
  // State for session data
  const [sessionData, setSessionData] = useState<CreateSessionResponse | null>(null);
  const [currentHighlightIndex, setCurrentHighlightIndex] = useState(0);
  const [userAnswer, setUserAnswer] = useState('');
  const [isProcessing, setIsProcessing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isSessionCompleted, setIsSessionCompleted] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  // Fetch session data when component mounts
  useEffect(() => {
    const fetchSessionData = async () => {
      if (!sessionId) {
        setError(t('errors.generic'));
        setIsLoading(false);
        return;
      }
      
      try {
        const sessionInfo: SessionData = await getSession(sessionId);
        
        // Convert SessionData to CreateSessionResponse format
        const convertedData: CreateSessionResponse = {
          session_id: sessionInfo.id,
          name: sessionInfo.name,
          total_highlights: sessionInfo.total_highlights,
          next_step: sessionInfo.next_step || {
            highlight_index: 0,
            highlight_text: '',
            question: t('session.yourAnswer')
          }
        };
        
        setSessionData(convertedData);
        setCurrentHighlightIndex(sessionInfo.next_step?.highlight_index || 0);
      } catch (err) {
        setError(err instanceof Error ? err.message : t('errors.generic'));
      } finally {
        setIsLoading(false);
      }
    };

    fetchSessionData();
  }, [sessionId]);

  const handleNext = async () => {
    if (!sessionId || !sessionData) return;

    setIsProcessing(true);
    setError(null);

    try {
      const request: ProcessAnswerRequest = {
        highlight_index: currentHighlightIndex,
        user_answer: userAnswer,
      };

      const response: ProcessAnswerResponse = await processAnswer(sessionId, request);

      if (response.status === 'completed') {
        setIsSessionCompleted(true);
        // Refresh the session list to update the status
        if (refreshSessionListRef.current) {
          refreshSessionListRef.current();
        }
      } else if (response.next_step) {
        // Update state with next step
        setCurrentHighlightIndex(response.next_step.highlight_index);
        setUserAnswer('');
        
        // Update session data with next step info
        setSessionData({
          ...sessionData,
          next_step: response.next_step
        });
      }
    } catch (err) {
      // Check if this is a 503 error and show a user-friendly message
      if (err instanceof Error && err.message.includes("LLM service unavailable")) {
        setError(t('errors.llmUnavailable'));
      } else {
        setError(err instanceof Error ? err.message : t('errors.generic'));
      }
    } finally {
      setIsProcessing(false);
    }
  };

  const handleSkip = () => {
    // For now, we'll just treat skip as submitting an empty answer
    setUserAnswer('');
    handleNext();
  };

  const handleRegenerateQuestion = async () => {
    if (!sessionId || !sessionData) return;

    setIsProcessing(true);
    setError(null);

    try {
      const request: RegenerateQuestionRequest = {
        highlight_index: currentHighlightIndex,
      };

      const response = await regenerateQuestion(sessionId, request);
      
      // Update the question in the session data
      if (sessionData.next_step) {
        setSessionData({
          ...sessionData,
          next_step: {
            ...sessionData.next_step,
            question: response.new_question
          }
        });
      }
    } catch (err) {
      // Check if this is a 503 error and show a user-friendly message
      if (err instanceof Error && err.message.includes("LLM service unavailable")) {
        setError(t('errors.llmUnavailable'));
      } else {
        setError(err instanceof Error ? err.message : t('errors.generic'));
      }
    } finally {
      setIsProcessing(false);
    }
  };

  const handleNewSession = () => {
    navigate('/');
  };

  if (isLoading) {
    return (
      <TwoPanelLayout onSessionListRefresh={() => { refreshSessionListRef.current = () => window.location.reload(); }}>
        <div className="session-page">
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
        <div className="session-page">
          <div className="error-container">
            <h2>{t('common.error')}</h2>
            <p>{error}</p>
            <button onClick={() => navigate('/')}>{t('common.back')}</button>
          </div>
        </div>
      </TwoPanelLayout>
    );
  }

  if (isSessionCompleted) {
    return (
      <TwoPanelLayout onSessionListRefresh={() => { refreshSessionListRef.current = () => window.location.reload(); }}>
        <div className="session-page">
          <div className="completion-container">
            <h2>{t('common.sessionCompleted')}</h2>
            <p>{t('session.sessionCompletedMessage')}</p>
            <div className="completion-actions">
              <button onClick={handleNewSession}>{t('common.startNewSession')}</button>
            </div>
          </div>
        </div>
      </TwoPanelLayout>
    );
  }

  if (!sessionData) {
    return (
      <TwoPanelLayout onSessionListRefresh={() => { refreshSessionListRef.current = () => window.location.reload(); }}>
        <div className="session-page">
          <div className="error-container">
            <h2>{t('common.error')}</h2>
            <p>{t('session.noSessionData')}</p>
            <button onClick={() => navigate('/')} className="secondary-btn">{t('common.back')}</button>
          </div>
        </div>
      </TwoPanelLayout>
    );
  }

  const progress = `${currentHighlightIndex + 1} ${t('session.of')} ${sessionData.total_highlights}`;

  return (
    <TwoPanelLayout onSessionListRefresh={() => { refreshSessionListRef.current = () => window.location.reload(); }}>
      <div className="session-page">
        <header className="session-header">
          <h1>{sessionData.name}</h1>
          <div className="progress-indicator">{progress}</div>
        </header>

        <main className="session-main">
          <div className="highlight-container">
            <h2>{t('session.highlight')}</h2>
            <div className="highlight-text">
              {sessionData.next_step?.highlight_text || t('session.noHighlightText')}
            </div>
          </div>

          <div className="question-container">
            <h2>{t('session.question')}</h2>
            <div className="question-text">
              {sessionData.next_step?.question || t('session.noQuestion')}
            </div>
          </div>

          <div className="answer-container">
            <h2>{t('session.yourAnswer')}</h2>
            <textarea
              value={userAnswer}
              onChange={(e) => setUserAnswer(e.target.value)}
              placeholder={t('session.answerPlaceholder')}
              rows={5}
              disabled={isProcessing}
            />
          </div>

          {error && <div className="error-message">{error}</div>}

          <div className="session-actions">
            <button 
              onClick={handleSkip} 
              disabled={isProcessing}
              className="secondary-btn"
            >
              {isProcessing ? <div className="small-spinner"></div> : t('common.skip')}
            </button>
            <button 
              onClick={handleRegenerateQuestion} 
              disabled={isProcessing}
              className="secondary-btn"
            >
              {isProcessing ? <div className="small-spinner"></div> : t('session.regenerateQuestion')}
            </button>
            <button 
              onClick={handleNext} 
              disabled={isProcessing}
            >
              {isProcessing ? <div className="small-spinner"></div> : t('common.next')}
            </button>
          </div>
        </main>
      </div>
    </TwoPanelLayout>
  );
};

export default SessionPage;