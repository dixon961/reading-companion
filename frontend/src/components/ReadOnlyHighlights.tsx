import React from 'react';
import type { SessionContent } from '../api/session';
import { useLanguage } from '../i18n/LanguageContext';

interface ReadOnlyHighlightsProps {
  content: SessionContent;
}

const ReadOnlyHighlights: React.FC<ReadOnlyHighlightsProps> = ({ content }) => {
  const { t, language } = useLanguage();
  
  return (
    <div className="json-renderer">
      <header className="json-header">
        <h1>{content.session.name}</h1>
        <p className="session-date">
          {t('review.sessionDate')}: {new Date(content.session.created_at).toLocaleDateString(language === 'ru' ? 'ru-RU' : 'en-US')}
        </p>
      </header>
      
      <div className="content-separator"></div>
      
      <div className="answers-container">
        {content.highlights.map((highlight, index) => (
          <div key={index} className="answer-block">
            {highlight.answered ? (
              <div className="answer-text compact">
                {highlight.answer}
              </div>
            ) : (
              <div className="unanswered-text compact">
                {t('review.noAnswerProvided')}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};

export default ReadOnlyHighlights;