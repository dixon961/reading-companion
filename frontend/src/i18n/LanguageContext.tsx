import React, { createContext, useContext, useState, useEffect } from 'react';
import type { ReactNode } from 'react';

// Define types for our translations
interface Translations {
  [key: string]: string | Translations;
}

// Define the context type
interface LanguageContextType {
  language: string;
  setLanguage: (language: string) => void;
  t: (key: string) => string;
  translations: Translations;
}

// Create the context with default values
const LanguageContext = createContext<LanguageContextType | undefined>(undefined);

// Define the provider props
interface LanguageProviderProps {
  children: ReactNode;
}

// Load translations function
const loadTranslations = async (language: string): Promise<Translations> => {
  try {
    // Use a switch statement to avoid dynamic imports that cause build issues
    switch (language) {
      case 'en':
        const enTranslations = await import('./en.json');
        return enTranslations.default;
      case 'ru':
      default:
        const ruTranslations = await import('./ru.json');
        return ruTranslations.default;
    }
  } catch (error) {
    console.warn(`Failed to load translations for ${language}, falling back to Russian`);
    const defaultTranslations = await import('./ru.json');
    return defaultTranslations.default;
  }
};

// Helper function to get nested translation
const getNestedTranslation = (obj: Translations, path: string): string => {
  const keys = path.split('.');
  let current: string | Translations = obj;
  
  for (const key of keys) {
    if (typeof current === 'object' && current !== null && key in current) {
      current = (current as Translations)[key];
    } else {
      return path; // Return the key if translation not found
    }
  }
  
  return typeof current === 'string' ? current : path;
};

// Language Provider Component
export const LanguageProvider: React.FC<LanguageProviderProps> = ({ children }) => {
  const [language, setLanguage] = useState<string>('ru');
  const [translations, setTranslations] = useState<Translations>({});

  // Load saved language from localStorage on initial load
  useEffect(() => {
    const savedLanguage = localStorage.getItem('language');
    if (savedLanguage && ['ru', 'en'].includes(savedLanguage)) {
      setLanguage(savedLanguage);
    }
  }, []);

  // Load translations when language changes
  useEffect(() => {
    const loadLanguage = async () => {
      try {
        const loadedTranslations = await loadTranslations(language);
        setTranslations(loadedTranslations);
      } catch (error) {
        console.error('Failed to load translations:', error);
        // Fallback to Russian if loading fails
        const ruTranslations = await import('./ru.json');
        setTranslations(ruTranslations.default);
        setLanguage('ru');
      }
    };

    loadLanguage();
  }, [language]);

  // Save language to localStorage when it changes
  useEffect(() => {
    localStorage.setItem('language', language);
  }, [language]);

  // Translation function
  const t = (key: string): string => {
    return getNestedTranslation(translations, key);
  };

  return (
    <LanguageContext.Provider value={{ language, setLanguage, t, translations }}>
      {children}
    </LanguageContext.Provider>
  );
};

// Custom hook to use the language context
export const useLanguage = (): LanguageContextType => {
  const context = useContext(LanguageContext);
  if (context === undefined) {
    throw new Error('useLanguage must be used within a LanguageProvider');
  }
  return context;
};

export default LanguageContext;