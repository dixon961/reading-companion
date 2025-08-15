import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Modal from '../components/Modal';
import TwoPanelLayout from '../components/TwoPanelLayout';
import { createSession } from '../api/session';
import { useLanguage } from '../i18n/LanguageContext';
import CustomFileInput from '../components/CustomFileInput';

const HomePage: React.FC = () => {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [file, setFile] = useState<File | null>(null);
  const [sessionName, setSessionName] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const navigate = useNavigate();
  const { t } = useLanguage();

  const openModal = () => setIsModalOpen(true);
  const closeModal = () => {
    setIsModalOpen(false);
    setFile(null);
    setSessionName('');
    setError(null);
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setFile(e.target.files[0]);
      setError(null);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!file) {
      setError(t('home.uploadFile'));
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const response = await createSession(file, sessionName);
      navigate(`/session/${response.session_id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('errors.generic'));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <TwoPanelLayout onSessionListRefresh={() => {}}>
      <div className="home-page">
        <header className="home-header">
          <h1>{t('home.title')}</h1>
          <p>{t('home.subtitle')}</p>
        </header>
        
        <main className="home-main">
          <button className="start-session-btn" onClick={openModal}>
            {t('home.startNewSession')}
          </button>
        </main>

        <Modal isOpen={isModalOpen} onClose={closeModal}>
          <div className="session-form">
            <h2>{t('home.createSession')}</h2>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label htmlFor="file">{t('home.uploadFile')}</label>
                <CustomFileInput
                  id="file"
                  accept=".txt"
                  onChange={handleFileChange}
                  required
                />
                <p className="help-text">{t('home.supportedFormats')}</p>
              </div>

              <div className="form-group">
                <label htmlFor="sessionName">{t('home.sessionName')}</label>
                <input
                  type="text"
                  id="sessionName"
                  value={sessionName}
                  onChange={(e) => setSessionName(e.target.value)}
                  placeholder={t('home.sessionNamePlaceholder')}
                />
              </div>

              {error && <div className="error-message">{error}</div>}

              <div className="form-actions">
                <button type="button" onClick={closeModal} disabled={isLoading}>
                  {t('common.cancel')}
                </button>
                <button type="submit" disabled={isLoading}>
                  {isLoading ? <div className="small-spinner"></div> : t('home.createSession')}
                </button>
              </div>
            </form>
          </div>
        </Modal>
      </div>
    </TwoPanelLayout>
  );
};

export default HomePage;