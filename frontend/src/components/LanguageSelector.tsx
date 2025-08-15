import React from 'react';
import { useLanguage } from '../i18n/LanguageContext';

interface LanguageSelectorProps {
  className?: string;
}

const LanguageSelector: React.FC<LanguageSelectorProps> = ({ className = '' }) => {
  const { language, setLanguage } = useLanguage();

  const handleLanguageChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    setLanguage(event.target.value);
  };

  return (
    <div className={`language-selector ${className}`}>
      <select 
        value={language} 
        onChange={handleLanguageChange}
        className="language-select"
      >
        <option value="ru">Русский</option>
        <option value="en">English</option>
      </select>
    </div>
  );
};

export default LanguageSelector;