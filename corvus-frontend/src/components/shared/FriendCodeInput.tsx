import { useState } from "react";
import { useFriendCode } from "../../hooks/useFriendCode";
import { validateFriendCode } from "../../api/deployments";

export default function FriendCodeInput() {
  const { friendCode, setFriendCode, clearFriendCode, hasFriendCode } = useFriendCode();
  const [inputValue, setInputValue] = useState("");
  const [error, setError] = useState("");
  const [isValidating, setIsValidating] = useState(false);

  const handleApply = async () => {
    if (!inputValue.trim()) return;
    setIsValidating(true); setError("");
    try {
      const result = await validateFriendCode(inputValue.trim());
      if (result.valid) { setFriendCode(inputValue.trim()); setInputValue(""); }
      else { setError("Invalid code"); }
    } catch { setFriendCode(inputValue.trim()); setInputValue(""); }
    finally { setIsValidating(false); }
  };

  if (hasFriendCode) {
    return (
      <div className="flex items-center gap-2" style={{ fontSize: "0.8rem" }}>
        <span style={{
          padding: "0.2rem 0.5rem",
          background: "var(--leaf-bg)",
          border: "1px solid var(--leaf)",
          color: "var(--leaf)",
          borderRadius: "1px",
          fontFamily: '"EB Garamond", serif',
          fontWeight: 700,
        }}>
          Extended access
          <svg width="12" height="12" viewBox="0 0 14 14" fill="none" style={{ display: "inline-block", verticalAlign: "middle", marginLeft: "4px" }}>
            <path d="M2 7.5L5.5 11L12 3" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
        </span>
        <button onClick={clearFriendCode} className="cursor-pointer transition-colors"
          style={{ color: "var(--sumi-ghost)" }}
          onMouseEnter={(e) => (e.currentTarget.style.color = "var(--sumi)")}
          onMouseLeave={(e) => (e.currentTarget.style.color = "var(--sumi-ghost)")}
          title={`Remove code: ${friendCode}`}
        >
          <svg width="12" height="12" viewBox="0 0 14 14" fill="none">
            <path d="M3 3L11 11M11 3L3 11" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
          </svg>
        </button>
      </div>
    );
  }

  return (
    <div className="flex items-center gap-2" style={{ fontSize: "0.8rem" }}>
      <span className="hidden sm:inline" style={{ color: "var(--sumi-wash)", fontStyle: "italic" }}>
        Access code?
      </span>
      <input type="text" value={inputValue}
        onChange={(e) => { setInputValue(e.target.value); setError(""); }}
        onKeyDown={(e) => e.key === "Enter" && handleApply()}
        placeholder="Enter code"
        className="ink-input"
        style={{ width: "6rem", padding: "0.2rem 0.5rem", fontSize: "0.8rem", borderBottomWidth: "1.5px", height: "auto" }}
      />
      <button onClick={handleApply} disabled={isValidating || !inputValue.trim()}
        className="ink-btn" style={{ padding: "0.2rem 0.6rem", fontSize: "0.8rem" }}>
        {isValidating ? "..." : "Apply"}
      </button>
      {error && <span style={{ color: "var(--vermillion)", fontSize: "0.75rem" }}>{error}</span>}
    </div>
  );
}
