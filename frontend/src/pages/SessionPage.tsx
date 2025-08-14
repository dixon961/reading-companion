import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { processAnswer, getSession, regenerateQuestion } from '../api/session';
import TwoPanelLayout from '../components/TwoPanelLayout';
import type { CreateSessionResponse, ProcessAnswerRequest, ProcessAnswerResponse, SessionData, RegenerateQuestionRequest } from '../api/session';

const SessionPage: React.FC = () => {
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();
  
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
        setError('No session ID provided');
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
            question: 'What are your thoughts on this highlight?'
          }
        };
        
        setSessionData(convertedData);
        setCurrentHighlightIndex(sessionInfo.next_step?.highlight_index || 0);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load session data');
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
      setError(err instanceof Error ? err.message : 'Failed to process answer');
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
      setError(err instanceof Error ? err.message : 'Failed to regenerate question');
    } finally {
      setIsProcessing(false);
    }
  };

  const handleNewSession = () => {
    navigate('/');
  };

  if (isLoading) {
    return (
      <TwoPanelLayout>
        <div className="session-page">
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
        <div className="session-page">
          <div className="error-container">
            <h2>Error</h2>
            <p>{error}</p>
            <button onClick={() => navigate('/')}>Back to Home</button>
          </div>
        </div>
      </TwoPanelLayout>
    );
  }

  if (isSessionCompleted) {
    return (
      <TwoPanelLayout>
        <div className="session-page">
          <div className="completion-container">
            <h2>Session Completed!</h2>
            <p>Your reading session has been successfully completed.</p>
            <div className="completion-actions">
              <button onClick={handleNewSession}>Start New Session</button>
            </div>
          </div>
        </div>
      </TwoPanelLayout>
    );
  }

  if (!sessionData) {
    return (
      <TwoPanelLayout>
        <div className="session-page">
          <div className="error-container">
            <h2>Error</h2>
            <p>No session data available</p>
            <button onClick={() => navigate('/')}>Back to Home</button>
          </div>
        </div>
      </TwoPanelLayout>
    );
  }

  const progress = `${currentHighlightIndex + 1} of ${sessionData.total_highlights}`;

  return (
    <TwoPanelLayout>
      <div className="session-page">
        <header className="session-header">
          <h1>{sessionData.name}</h1>
          <div className="progress-indicator">{progress}</div>
        </header>

        <main className="session-main">
          <div className="highlight-container">
            <h2>Highlight</h2>
            <div className="highlight-text">
              {sessionData.next_step?.highlight_text || 'No highlight text available'}
            </div>
          </div>

          <div className="question-container">
            <h2>Question</h2>
            <div className="question-text">
              {sessionData.next_step?.question || 'No question available'}
            </div>
          </div>

          <div className="answer-container">
            <h2>Your Answer</h2>
            <textarea
              value={userAnswer}
              onChange={(e) => setUserAnswer(e.target.value)}
              placeholder="Type your thoughts here..."
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
              {isProcessing ? <div className="small-spinner"></div> : 'Skip'}
            </button>
            <button 
              onClick={handleRegenerateQuestion} 
              disabled={isProcessing}
              className="secondary-btn"
            >
              {isProcessing ? <div className="small-spinner"></div> : 'Regenerate Question'}
            </button>
            <button 
              onClick={handleNext} 
              disabled={isProcessing}
            >
              {isProcessing ? <div className="small-spinner"></div> : 'Next'}
            </button>
          </div>
        </main>
      </div>
    </TwoPanelLayout>
  );
};

export default SessionPage;