import React, { useState } from 'react';
import type { SessionMetadata } from '../api/session';
import { useNavigate } from 'react-router-dom';

interface SessionListItemProps {
  session: SessionMetadata;
  onRename: (sessionId: string, newName: string) => Promise<void>;
  onDelete: (sessionId: string) => Promise<void>;
}

const SessionListItem: React.FC<SessionListItemProps> = ({ session, onRename, onDelete }) => {
  const navigate = useNavigate();
  const [isEditing, setIsEditing] = useState(false);
  const [editName, setEditName] = useState(session.name);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'status-completed';
      case 'in_progress':
        return 'status-in-progress';
      default:
        return 'status-default';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'completed':
        return 'Завершена';
      case 'in_progress':
        return 'В процессе';
      default:
        return status;
    }
  };

  const handleClick = () => {
    if (session.status === 'completed') {
      navigate(`/review/${session.id}`);
    } else {
      navigate(`/session/${session.id}`);
    }
  };

  const handleRenameClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    setIsEditing(true);
    setEditName(session.name);
  };

  const handleRenameSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    e.stopPropagation();
    
    if (editName.trim() === '' || editName === session.name) {
      setIsEditing(false);
      return;
    }
    
    setIsProcessing(true);
    try {
      await onRename(session.id, editName.trim());
      setIsEditing(false);
    } catch (err) {
      alert('Failed to rename session: ' + (err instanceof Error ? err.message : 'Unknown error'));
    } finally {
      setIsProcessing(false);
    }
  };

  const handleDeleteClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    setShowDeleteConfirm(true);
  };

  const handleDeleteConfirm = async () => {
    setIsProcessing(true);
    try {
      await onDelete(session.id);
      setShowDeleteConfirm(false);
    } catch (err) {
      alert('Failed to delete session: ' + (err instanceof Error ? err.message : 'Unknown error'));
      setIsProcessing(false);
    }
  };

  const handleCancelDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    setShowDeleteConfirm(false);
  };

  if (showDeleteConfirm) {
    return (
      <div className="session-list-item" onClick={(e) => e.stopPropagation()}>
        <div className="delete-confirm-overlay">
          <div className="delete-confirm-content">
            <h3>Удалить сессию?</h3>
            <p>Вы уверены, что хотите удалить сессию "{session.name}"? Это действие нельзя отменить.</p>
            <div className="delete-confirm-actions">
              <button 
                onClick={handleCancelDelete} 
                disabled={isProcessing}
                className="secondary-btn"
              >
                Отмена
              </button>
              <button 
                onClick={handleDeleteConfirm} 
                disabled={isProcessing}
                className="danger-btn"
              >
                {isProcessing ? 'Удаление...' : 'Удалить'}
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (isEditing) {
    return (
      <div className="session-list-item" onClick={(e) => e.stopPropagation()}>
        <form onSubmit={handleRenameSubmit} className="rename-form">
          <input
            type="text"
            value={editName}
            onChange={(e) => setEditName(e.target.value)}
            autoFocus
            disabled={isProcessing}
          />
          <div className="rename-actions">
            <button 
              type="button" 
              onClick={() => setIsEditing(false)} 
              disabled={isProcessing}
              className="secondary-btn"
            >
              Отмена
            </button>
            <button 
              type="submit" 
              disabled={isProcessing || editName.trim() === '' || editName === session.name}
            >
              {isProcessing ? 'Сохранение...' : 'Сохранить'}
            </button>
          </div>
        </form>
      </div>
    );
  }

  return (
    <div className="session-list-item" onClick={handleClick}>
      <div className="session-info">
        <h3 className="session-name">{session.name}</h3>
        <div className="session-meta">
          <span className={`session-status ${getStatusColor(session.status)}`}>
            {getStatusText(session.status)}
          </span>
          <span className="session-date">{formatDate(session.created_at)}</span>
        </div>
      </div>
      <div className="session-actions">
        <button 
          className="icon-button" 
          aria-label="Переименовать"
          onClick={handleRenameClick}
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"></path>
          </svg>
        </button>
        <button 
          className="icon-button" 
          aria-label="Удалить"
          onClick={handleDeleteClick}
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <polyline points="3 6 5 6 21 6"></polyline>
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
          </svg>
        </button>
      </div>
    </div>
  );
};

export default SessionListItem;