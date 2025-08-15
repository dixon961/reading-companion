import React, { useRef } from 'react';
import { useLanguage } from '../i18n/LanguageContext';

interface CustomFileInputProps {
  id: string;
  accept: string;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  required?: boolean;
}

const CustomFileInput: React.FC<CustomFileInputProps> = ({ 
  id, 
  accept, 
  onChange, 
  required 
}) => {
  const { t } = useLanguage();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [fileName, setFileName] = React.useState<string | null>(null);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setFileName(e.target.files[0].name);
    } else {
      setFileName(null);
    }
    onChange(e);
  };

  const handleButtonClick = () => {
    if (fileInputRef.current) {
      fileInputRef.current.click();
    }
  };

  return (
    <div className="custom-file-input">
      <input
        type="file"
        id={id}
        accept={accept}
        onChange={handleFileChange}
        required={required}
        ref={fileInputRef}
        style={{ display: 'none' }}
      />
      <button 
        type="button" 
        onClick={handleButtonClick}
        className="file-input-button"
      >
        {t('home.chooseFile')}
      </button>
      <span className="file-input-text">
        {fileName || t('home.noFileChosen')}
      </span>
    </div>
  );
};

export default CustomFileInput;