import React, { createContext, useContext, useState, useEffect } from 'react';
import { 
  signInWithEmailAndPassword, 
  createUserWithEmailAndPassword,
  signOut,
  onAuthStateChanged,
  GoogleAuthProvider,
  signInWithPopup
} from 'firebase/auth';
import { auth } from '../config/firebase';

const AuthContext = createContext({});

export const useAuth = () => {
  return useContext(AuthContext);
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!auth) {
      setLoading(false);
      return;
    }
    
    const unsubscribe = onAuthStateChanged(auth, (user) => {
      setUser(user);
      setLoading(false);
    });

    return unsubscribe;
  }, []);

  const signup = (email, password) => {
    if (!auth) {
      return Promise.reject(new Error('Authentication not configured'));
    }
    return createUserWithEmailAndPassword(auth, email, password);
  };

  const login = (email, password) => {
    if (!auth) {
      return Promise.reject(new Error('Authentication not configured'));
    }
    return signInWithEmailAndPassword(auth, email, password);
  };

  const loginWithGoogle = () => {
    if (!auth) {
      return Promise.reject(new Error('Authentication not configured'));
    }
    const provider = new GoogleAuthProvider();
    return signInWithPopup(auth, provider);
  };

  const logout = () => {
    if (!auth) {
      return Promise.resolve();
    }
    return signOut(auth);
  };

  const value = {
    user,
    loading,
    signup,
    login,
    loginWithGoogle,
    logout,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};
