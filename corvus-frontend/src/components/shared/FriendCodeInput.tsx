import { useState } from "react";
import { useFriendCode } from "../../hooks/useFriendCode";
import { validateFriendCode } from "../../api/deployments";

/** Small friend code input widget for the header */
export default function FriendCodeInput() {
  const { friendCode, setFriendCode, clearFriendCode, hasFriendCode } =
    useFriendCode();
  const [inputValue, setInputValue] = useState("");
  const [error, setError] = useState("");
  const [isValidating, setIsValidating] = useState(false);

  const handleApply = async () => {
    if (!inputValue.trim()) return;
    setIsValidating(true);
    setError("");
    try {
      const result = await validateFriendCode(inputValue.trim());
      if (result.valid) {
        setFriendCode(inputValue.trim());
        setInputValue("");
      } else {
        setError("Invalid code");
      }
    } catch {
      // If validate endpoint doesn't exist, just store the code
      setFriendCode(inputValue.trim());
      setInputValue("");
    } finally {
      setIsValidating(false);
    }
  };

  if (hasFriendCode) {
    return (
      <div className="flex items-center gap-2 text-xs">
        <span className="px-2 py-1 bg-gray-100 border border-gray-300 rounded text-gray-700">
          Extended access ✓
        </span>
        <button
          onClick={clearFriendCode}
          className="text-gray-400 hover:text-gray-600 cursor-pointer"
          title={`Remove code: ${friendCode}`}
        >
          ×
        </button>
      </div>
    );
  }

  return (
    <div className="flex items-center gap-2 text-xs">
      <span className="text-gray-500 hidden sm:inline">Access code?</span>
      <input
        type="text"
        value={inputValue}
        onChange={(e) => {
          setInputValue(e.target.value);
          setError("");
        }}
        onKeyDown={(e) => e.key === "Enter" && handleApply()}
        placeholder="Enter code"
        className="px-2 py-1 border border-gray-300 rounded text-xs w-24 focus:outline-none focus:border-black"
      />
      <button
        onClick={handleApply}
        disabled={isValidating || !inputValue.trim()}
        className="px-2 py-1 bg-black text-white rounded text-xs hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
      >
        {isValidating ? "..." : "Apply"}
      </button>
      {error && <span className="text-red-500">{error}</span>}
    </div>
  );
}

