import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Modal from '../components/Modal';
import { createSession } from '../api/session';

const HomePage: React.FC = () => {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [file, setFile] = useState<File | null>(null);
  const [sessionName, setSessionName] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const navigate = useNavigate();

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
      setError('Please select a file');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const response = await createSession(file, sessionName);
      navigate(`/session/${response.session_id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create session');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="home-page">
      <header className="home-header">
        <h1>Interactive Reading Companion</h1>
        <p>Transform your book highlights into meaningful insights</p>
      </header>
      
      <main className="home-main">
        <button className="start-session-btn" onClick={openModal}>
          Start New Session
        </button>
      </main>

      <Modal isOpen={isModalOpen} onClose={closeModal}>
        <div className="session-form">
          <h2>Create New Session</h2>
          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <label htmlFor="file">Upload Highlights File (.txt)</label>
              <input
                type="file"
                id="file"
                accept=".txt"
                onChange={handleFileChange}
                required
              />
              <p className="help-text">Supported: Kon-Tiki 2 export format</p>
            </div>

            <div className="form-group">
              <label htmlFor="sessionName">Session Name (Optional)</label>
              <input
                type="text"
                id="sessionName"
                value={sessionName}
                onChange={(e) => setSessionName(e.target.value)}
                placeholder="Leave blank to auto-generate"
              />
            </div>

            {error && <div className="error-message">{error}</div>}

            <div className="form-actions">
              <button type="button" onClick={closeModal} disabled={isLoading}>
                Cancel
              </button>
              <button type="submit" disabled={isLoading}>
                {isLoading ? 'Creating...' : 'Create Session'}
              </button>
            </div>
          </form>
        </div>
      </Modal>
    </div>
  );
};

export default HomePage;