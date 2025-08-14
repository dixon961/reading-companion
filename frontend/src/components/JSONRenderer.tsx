import React from 'react';
import type { SessionContent } from '../api/session';

interface JSONRendererProps {
  content: SessionContent;
}

const JSONRenderer: React.FC<JSONRendererProps> = ({ content }) => {
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

export default JSONRenderer;