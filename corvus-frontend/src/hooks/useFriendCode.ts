import { useState, useCallback } from "react";
import { STORAGE_KEY_FRIEND_CODE } from "../config/constants";

export function useFriendCode() {
  const [friendCode, setFriendCodeState] = useState<string | null>(() => {
    try {
      return localStorage.getItem(STORAGE_KEY_FRIEND_CODE);
    } catch {
      return null;
    }
  });

  const setFriendCode = useCallback((code: string) => {
    localStorage.setItem(STORAGE_KEY_FRIEND_CODE, code);
    setFriendCodeState(code);
  }, []);

  const clearFriendCode = useCallback(() => {
    localStorage.removeItem(STORAGE_KEY_FRIEND_CODE);
    setFriendCodeState(null);
  }, []);

  return {
    friendCode,
    setFriendCode,
    clearFriendCode,
    hasFriendCode: friendCode !== null && friendCode !== "",
  };
}

