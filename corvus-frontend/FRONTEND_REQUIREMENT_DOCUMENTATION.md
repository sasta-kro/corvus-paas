

# Corvus PaaS Frontend — Requirements & Implementation Document

---

## 1. Project Overview

### 1.1 What This Is

The frontend for Corvus PaaS is a public-facing demo platform that lets anyone deploy a static website and get a live URL. It is not a developer dashboard — it is a showcase. The primary audience is employers, recruiters, and curious visitors who land on `corvus.sasta.dev` and want to see what this thing does.

The frontend is built for functional correctness, component structure, and full API integration. A professional web developer will later customize the visual design, typography, colors, and animations. The implementation therefore uses a simple black-and-white theme with base system fonts, clean component boundaries, and Tailwind utility classes that are easy to remap.

### 1.2 Hosting

The frontend is a static React app hosted on Vercel or Netlify (free tier). It communicates with the Go backend via `fetch()` over HTTPS through a Cloudflare Tunnel. The backend runs on a Fedora VM on a home lab server.

### 1.3 Handoff Contract with Web Developer

The web developer receives:
- A fully functional React app with all pages, components, API calls, state management, and interactive logic working
- Black-and-white theme with Tailwind utility classes on every element
- Base system font stack (`font-sans` default)
- No animations beyond functional transitions (e.g., conditional rendering)
- Clean component tree where every visual element is its own component file

The web developer modifies:
- Color palette (replace black/white/gray values across Tailwind config and classes)
- Typography (font family, sizes, weights, line heights)
- Spacing and visual rhythm
- Animations and transitions (Framer Motion or CSS)
- Background effects (gradient mesh, particles, or none)
- Logo integration
- Hover states, focus rings, micro-interactions
- Responsive breakpoints and mobile layout
- Overall visual polish

The web developer should NOT need to modify:
- Component structure or hierarchy
- State management logic
- API call functions or data flow
- Routing
- Business logic (TTL calculation, session enforcement, validation)

---

## 2. Pages & Routing

### 2.1 Route Table

| Path | Component | Purpose |
|---|---|---|
| `/` | `LandingPage` | Hero + deploy panel + active deployment view |
| `/d/:slug` | `DeploymentViewerPage` | Single deployment status, link, countdown, actions |

React Router v6 handles routing. A catch-all `*` route redirects to `/`.

### 2.2 Page 1: Landing Page (`/`)

This is the main experience. Layout from top to bottom:

**Section A: Header Bar**
- Logo placeholder (left): an `<img>` tag with `src` pointing to a placeholder image (e.g., a simple bird silhouette SVG or just the text "Corvus"). The web developer replaces this with the final logo.
- "Corvus" text next to the logo
- Right side: small unobtrusive "Friend Code" input. A text field + apply button. When a valid code is applied, a subtle indicator shows "Extended access" or similar. The code is stored in `localStorage` and sent with every create deployment request.

**Section B: Hero**
- Large heading: "Deploy a website in seconds."
- Subheading: one line explaining what this is. Example: "A self-hosted PaaS platform. Pick a preset, upload a zip, or paste a GitHub URL."
- Text is centered. Black text on white background (or white on black if the web dev goes dark mode — the implementation uses neutral Tailwind classes that flip easily).

**Section C: Deploy Panel**
- This is the core interactive area. Described in detail in Section 3.
- Takes up the visual center of the page.
- Has three tabs: "Quick Deploy", "Zip Upload", "GitHub Repo"

**Section D: Active Deployment View**
- Hidden by default. Appears below the deploy panel (or replaces it) when the user has an active deployment.
- Shows: status, live URL, countdown timer, redeploy button, delete button.
- Described in detail in Section 5.

**Section E: Footer**
- Minimal. Links to: GitHub repo, your GitHub profile, your LinkedIn.
- "Built with Go, Docker, Traefik" or similar tech credits.
- Small text.

### 2.3 Page 2: Deployment Viewer (`/d/:slug`)

A standalone page for viewing a single deployment. Users reach this by:
- Being redirected here after deploying from the landing page
- Bookmarking or sharing the URL
- Clicking "View Deployment" from the landing page's active deployment section

**Layout:**
- Header bar (same as landing page, with logo and friend code)
- Centered card showing:
    - Deployment name
    - Status badge (deploying / live / failed / expired)
    - Live URL (clickable, opens in new tab) — only shown when status is live
    - Countdown timer — only shown when status is live
    - "Open Site" button — opens the deployed site in a new tab
    - "Copy Link" button — copies the deployment URL to clipboard
    - "Redeploy" button — triggers `POST /api/deployments/:uuid/redeploy`
    - "Delete" button — triggers `DELETE /api/deployments/:uuid`, then redirects to `/`
    - Source info: shows whether it was a zip upload, GitHub repo, or preset
    - Timestamps: created at, last updated
- If the deployment does not exist (404 from API), show a "Deployment not found" message with a link back to `/`.

---

## 3. Deploy Panel — Detailed Specification

### 3.1 Tab Structure

Three tabs rendered using a tab component (shadcn/ui Tabs or custom):

```
[ Quick Deploy ]  [ Zip Upload ]  [ GitHub Repo ]
```

"Quick Deploy" is the default selected tab.

### 3.2 Tab 1: Quick Deploy

Displays 4 preset cards in a 2x2 grid (or 4-column row on wide screens, stacking on mobile).

**Card 1: "Vite Starter"**
- Description: "A minimal Vite app. Deploys in seconds."
- Icon/emoji placeholder: ⚡
- On click: immediately sends create deployment request with preset config
- Preset config:
  ```
  name: "Vite Starter"
  source_type: "github"
  github_url: "https://github.com/sasta-kro/corvus-preset-vite-starter.git"
  branch: "main"
  build_command: "npm ci && npm run build"
  output_directory: "dist"
  ```

**Card 2: "React App"**
- Description: "A React + Vite template with hot reload."
- Icon/emoji placeholder: ⚛️
- On click: sends create deployment request with preset config
- Preset config:
  ```
  name: "React App"
  source_type: "github"
  github_url: "https://github.com/sasta-kro/corvus-preset-react-app.git"
  branch: "main"
  build_command: "npm ci && npm run build"
  output_directory: "dist"
  ```

**Card 3: "About Corvus"**
- Description: "A custom page about this platform with links to the source code."
- Icon/emoji placeholder: 🐦
- On click: sends create deployment request with preset config
- Preset config:
  ```
  name: "About Corvus"
  source_type: "github"
  github_url: "https://github.com/sasta-kro/corvus-preset-about.git"
  branch: "main"
  build_command: "npm ci && npm run build"
  output_directory: "dist"
  ```

**Card 4: "Your Message"**
- Description: "Create a page with your custom message."
- Icon/emoji placeholder: ✍️
- On click: does NOT immediately deploy. Instead, shows a text input modal/inline field:
    - Label: "What should your page say?"
    - Input: text field, max 100 characters, character counter shown
    - Submit button: "Deploy My Message"
    - On submit: sends create deployment request with the user's text as an environment variable
- Preset config:
  ```
  name: "Custom Message"
  source_type: "github"
  github_url: "https://github.com/sasta-kro/corvus-preset-message.git"
  branch: "main"
  build_command: "npm ci && npm run build"
  output_directory: "dist"
  environment_variables: { "VITE_CORVUS_MESSAGE": "<user input text>" }
  ```
  The preset repo reads `import.meta.env.VITE_CORVUS_MESSAGE` and displays it as a large heading.

**All presets:** When clicked (or after text input for Card 4), the deploy panel transitions to the progress view (Section 4). The user cannot click another preset while a deployment is in progress.

### 3.3 Tab 2: Zip Upload

**Layout:**
- Drag-and-drop zone: a dashed-border rectangle area
    - Default state: "Drag and drop a .zip file here, or click to browse"
    - Hover/dragover state: border color change, "Drop your file here"
    - File selected state: shows file name, file size, and a remove button (X)
- Below the drop zone:
    - "Output Directory" text input (default value: `.`, placeholder: `e.g., dist, build, .`)
    - "Build Command" text input (optional, placeholder: `e.g., npm ci && npm run build`)
- "Deploy" button: disabled until a file is selected
- File size validation:
    - If file > 50MB: show inline error "File exceeds the 50MB limit", disable deploy button
    - Show file size next to file name after selection

**Drag-and-drop behavior:**
- Accept only `.zip` files. If user drops a non-zip file, show inline error "Only .zip files are accepted"
- Use the HTML5 drag-and-drop API (`onDragOver`, `onDragEnter`, `onDragLeave`, `onDrop`)
- Also support clicking the zone to open a native file picker (`<input type="file" accept=".zip">` hidden, triggered by click)

**On deploy click:**
- Construct a `FormData` object:
  ```
  name: extracted from zip file name (e.g., "my-site.zip" → "my-site")
  source_type: "zip"
  file: the zip File object
  output_directory: value from input field
  build_command: value from input field (empty string if not provided)
  friend_code: from localStorage (if present)
  ```
- Send `POST /api/deployments` with `Content-Type: multipart/form-data`
- Transition to progress view

### 3.4 Tab 3: GitHub Repo

**Layout:**
- "Repository URL" text input (required, placeholder: `https://github.com/user/repo`)
- "Branch" text input (default value: `main`)
- "Build Command" text input (required for GitHub deploys, placeholder: `npm ci && npm run build`)
- "Output Directory" text input (default value: `dist`)
- "Deploy" button: disabled until URL and build command are filled

**Validation:**
- URL must start with `https://github.com/` — if not, show inline error "Only public GitHub repositories are supported"
- URL must end with `.git` or be a valid GitHub repo URL — append `.git` if missing before sending to backend
- Build command cannot be empty for GitHub deploys — show inline error "Build command is required for GitHub deployments"

**On deploy click:**
- Send `POST /api/deployments` as `multipart/form-data` (to match the backend's existing multipart parsing):
  ```
  name: extracted from repo URL (e.g., "https://github.com/user/my-app.git" → "my-app")
  source_type: "github"
  github_url: the URL value
  branch: value from input field
  build_command: value from input field
  output_directory: value from input field
  friend_code: from localStorage (if present)
  ```
- Transition to progress view

---

## 4. Deployment Progress View

### 4.1 When It Appears

The progress view replaces the deploy panel content (same area on the page) when:
- A preset card is clicked
- The "Deploy" button is clicked on zip or GitHub tabs

### 4.2 Layout

```
┌─────────────────────────────────────────────────┐
│                                                 │
│  Deploying "React App"                          │
│                                                 │
│  ✓  Deployment created                          │
│  ✓  Source files received                       │
│  ⟳  Building project...                        │
│  ○  Starting server                             │
│  ○  Going live                                  │
│                                                 │
└─────────────────────────────────────────────────┘
```

### 4.3 Step Definitions

Five visual steps. These are NOT mapped 1:1 to backend pipeline stages. They are cosmetic steps driven by the deployment status from polling.

| Step | Label | Triggered by |
|---|---|---|
| 1 | "Deployment created" | Immediately on receiving 201/202 response from create/redeploy API |
| 2 | "Source files received" | After 2 seconds (simulated delay for visual pacing) |
| 3 | "Building project..." | After 4 seconds OR if status is still `"deploying"` at any poll |
| 4 | "Starting server" | After 6 seconds AND status is still `"deploying"` |
| 5 | "Going live" | When polled status changes to `"live"` |

**Step states:**
- Completed: checkmark icon (✓), muted text color
- In progress: spinner icon (⟳), normal text, optionally with a pulsing dot
- Pending: empty circle (○), muted text
- Failed: X icon (✗), red text — shown on the step that was "in progress" when failure was detected

**If status becomes `"failed"`:** The currently in-progress step gets the failed state. All subsequent steps stay pending. Show an error message below the steps: "Deployment failed. Check the build command and try again." and a "Try Again" button that returns to the deploy panel.

**If status becomes `"live"`:** All steps become completed. After a brief moment (500ms), the progress view transitions to the active deployment view (Section 5).

### 4.4 Polling Logic

- Start polling `GET /api/deployments/:id` immediately after receiving the create response
- Poll interval: every 2 seconds
- Stop polling when status is `"live"` or `"failed"`
- Maximum poll duration: 120 seconds (2 minutes). If still `"deploying"` after 120 seconds, show a timeout message: "Deployment is taking longer than expected. It may still complete." with a "Keep Waiting" button (resets the 120s timer) and a "Cancel" button (deletes the deployment).

### 4.5 Zip Upload: Show Upload Progress

For zip uploads specifically, add an upload progress indicator between step 1 and step 2. Use `XMLHttpRequest` (not `fetch`) to get upload progress events, or use `fetch` with a progress wrapper. Show a percentage or progress bar during the upload. Step 2 ("Source files received") completes when the upload finishes and the 201 response is received.

---

## 5. Active Deployment View

### 5.1 When It Appears

- After the progress view completes successfully (status becomes `"live"`)
- On page load if `localStorage` has an active deployment ID and the deployment is still live

### 5.2 Layout

```
┌─────────────────────────────────────────────────┐
│                                                 │
│  🟢 Your site is live                           │
│                                                 │
│  "React App"                                    │
│                                                 │
│  https://tidal-ridge-c494.corvus.sasta.dev      │
│                                                 │
│  [ Open Site ]  [ Copy Link ]                   │
│                                                 │
│  Expires in 14:32                               │
│                                                 │
│  [ View Details ]  [ Delete ]                   │
│                                                 │
└─────────────────────────────────────────────────┘
```

### 5.3 Elements

**Status indicator:** Green circle + "Your site is live" text.

**Deployment name:** The name from the deployment record.

**Live URL:** Displayed as a clickable link. Opens in a new tab. The URL format depends on the domain setup — for now, use the `url` field from the API response directly.

**"Open Site" button:** Opens the deployment URL in a new tab (`window.open(url, '_blank')`).

**"Copy Link" button:** Copies the URL to clipboard via `navigator.clipboard.writeText()`. On success, button text briefly changes to "Copied!" for 2 seconds, then reverts.

**Countdown timer:**
- Calculated from `created_at + TTL`. The TTL value is known to the frontend (default: 15 minutes, friend code: extended TTL).
- Displays as `MM:SS` format, counting down every second.
- When it reaches `00:00`, the frontend polls the API one final time. If the deployment is gone (404), show "Deployment expired" and clear `localStorage`. If still alive (backend cleanup has slight delay), poll again in 5 seconds.
- The TTL duration should be a constant in the frontend config, not hardcoded in multiple places.

**"View Details" button:** Navigates to `/d/:slug`.

**"Delete" button:** Shows a confirmation dialog ("Are you sure? This will permanently delete your deployment."). On confirm, sends `DELETE /api/deployments/:uuid`. On success, clears `localStorage`, shows a brief "Deployment deleted" message, and returns to the deploy panel.

### 5.4 Session Enforcement

- When the landing page loads, check `localStorage` for `corvus_active_deployment_id` and `corvus_active_deployment_slug`.
- If found, poll `GET /api/deployments/:id`:
    - If response is 200 and status is `"live"`: show the active deployment view instead of the deploy panel. The user cannot create a new deployment until this one expires or is deleted.
    - If response is 200 and status is `"deploying"`: show the progress view.
    - If response is 200 and status is `"failed"`: clear `localStorage`, show the deploy panel with an error toast.
    - If response is 404: the deployment expired or was deleted. Clear `localStorage`, show the deploy panel.
- When a new deployment is created: store `{ id, slug }` in `localStorage`.
- When a deployment is deleted: clear `localStorage`.

---

## 6. Friend Code System

### 6.1 UI

A small input field in the header bar, right-aligned. Not prominent — this is not a core feature, it is a side channel for extended access.

**Default state:** Small text "Have an access code?" next to a short text input and an "Apply" button.

**After applying a valid code:** The input area is replaced with a small badge: "Extended access ✓". The code is stored in `localStorage` as `corvus_friend_code`.

**After applying an invalid code:** Inline error text: "Invalid code". The input stays visible.

**Validation:** The friend code is sent with the create deployment request. The backend returns a normal 201 response regardless of whether the code is valid (the code just affects TTL). The frontend does not know if the code is valid until deployment is created. Alternatively, add a simple `GET /api/validate-code?code=xyz` endpoint that returns `{ "valid": true/false }` so the frontend can give instant feedback. This is a one-line backend handler.

### 6.2 Storage

- `localStorage` key: `corvus_friend_code`
- Persists across page reloads and sessions
- Sent as a field in every create deployment request (if present)
- A "Clear" or "Remove" action next to the badge lets the user remove the code

---

## 7. Component Architecture

### 7.1 Component Tree

```
App
├── Header
│   ├── LogoPlaceholder
│   └── FriendCodeInput
│
├── Routes
│   ├── LandingPage
│   │   ├── HeroSection
│   │   ├── DeployPanel (conditionally rendered based on session state)
│   │   │   ├── DeployTabs
│   │   │   │   ├── QuickDeployTab
│   │   │   │   │   ├── PresetCard (x4)
│   │   │   │   │   └── MessageInputModal
│   │   │   │   ├── ZipUploadTab
│   │   │   │   │   ├── DragDropZone
│   │   │   │   │   └── ZipConfigFields
│   │   │   │   └── GitHubRepoTab
│   │   │   │       └── GitHubConfigFields
│   │   │   └── DeployButton
│   │   │
│   │   ├── DeployProgressView (conditionally rendered during deployment)
│   │   │   ├── ProgressStep (x5)
│   │   │   └── ErrorMessage (conditional)
│   │   │
│   │   ├── ActiveDeploymentView (conditionally rendered when deployment is live)
│   │   │   ├── StatusBadge
│   │   │   ├── LiveUrlDisplay
│   │   │   ├── CountdownTimer
│   │   │   ├── ActionButtons (Open, Copy, Redeploy, Delete)
│   │   │   └── DeleteConfirmDialog
│   │   │
│   │   └── Footer
│   │
│   └── DeploymentViewerPage
│       ├── DeploymentDetailCard
│       │   ├── StatusBadge
│       │   ├── LiveUrlDisplay
│       │   ├── CountdownTimer
│       │   ├── DeploymentMetadata
│       │   └── ActionButtons
│       └── NotFoundMessage (conditional)
│
└── Toasts/Notifications (global)
```

### 7.2 File Structure

```
src/
├── main.tsx                          # React entry point
├── App.tsx                           # Router setup, global providers
│
├── pages/
│   ├── LandingPage.tsx
│   └── DeploymentViewerPage.tsx
│
├── components/
│   ├── layout/
│   │   ├── Header.tsx
│   │   ├── Footer.tsx
│   │   └── LogoPlaceholder.tsx
│   │
│   ├── deploy/
│   │   ├── DeployPanel.tsx           # Container for tabs + state machine
│   │   ├── DeployTabs.tsx            # Tab navigation wrapper
│   │   ├── QuickDeployTab.tsx        # Preset cards grid
│   │   ├── PresetCard.tsx            # Individual preset card
│   │   ├── MessageInputModal.tsx     # Text input for "Your Message" preset
│   │   ├── ZipUploadTab.tsx          # Drag-drop + config fields
│   │   ├── DragDropZone.tsx          # Drag-and-drop area component
│   │   ├── GitHubRepoTab.tsx         # GitHub URL + config fields
│   │   └── DeployButton.tsx          # Shared deploy trigger button
│   │
│   ├── progress/
│   │   ├── DeployProgressView.tsx    # Progress steps container
│   │   └── ProgressStep.tsx          # Single step row (icon + label)
│   │
│   ├── deployment/
│   │   ├── ActiveDeploymentView.tsx  # Live deployment card on landing page
│   │   ├── DeploymentDetailCard.tsx  # Full detail card for viewer page
│   │   ├── StatusBadge.tsx           # deploying/live/failed badge
│   │   ├── LiveUrlDisplay.tsx        # URL with copy button
│   │   ├── CountdownTimer.tsx        # MM:SS countdown
│   │   └── DeleteConfirmDialog.tsx   # Confirmation modal
│   │
│   └── shared/
│       ├── FriendCodeInput.tsx       # Header friend code widget
│       └── Toast.tsx                 # Notification toast component
│
├── api/
│   ├── client.ts                     # Base fetch wrapper (base URL, error handling)
│   └── deployments.ts               # API functions: createDeployment, getDeployment, deleteDeployment, redeployDeployment
│
├── hooks/
│   ├── useDeploymentPolling.ts       # Polls GET /api/deployments/:id at interval
│   ├── useCountdown.ts              # Countdown timer logic
│   ├── useActiveDeployment.ts       # localStorage read/write for session state
│   └── useFriendCode.ts             # localStorage read/write for friend code
│
├── config/
│   └── constants.ts                  # API base URL, TTL duration, poll interval, max file size, presets config
│
├── types/
│   └── deployment.ts                 # TypeScript types matching backend API response shapes
│
└── lib/
    └── utils.ts                      # Utility functions (formatTime, extractRepoName, etc.)
```

---

## 8. API Integration Layer

### 8.1 Base Client (`api/client.ts`)

A thin wrapper around `fetch()` that:
- Prepends the API base URL (from `config/constants.ts`)
- Sets `Content-Type: application/json`



### 8.2 Deployment API Functions (`api/deployments.ts`)

Each function maps to one backend endpoint. These are the only functions that call the base client.

```typescript
// Creates a new deployment from a zip file upload
async function createZipDeployment(params: {
  file: File;
  outputDirectory: string;
  buildCommand: string;
  friendCode?: string;
}): Promise<Deployment>
// Constructs FormData with: name (from filename), source_type: "zip", file, output_directory, build_command, friend_code
// Calls apiPostFormData("/api/deployments", formData)

// Creates a new deployment from a GitHub repo URL
async function createGitHubDeployment(params: {
  name: string;
  githubUrl: string;
  branch: string;
  buildCommand: string;
  outputDirectory: string;
  environmentVariables?: Record<string, string>;
  friendCode?: string;
}): Promise<Deployment>
// Constructs FormData with all fields
// Calls apiPostFormData("/api/deployments", formData)

// Fetches a single deployment by ID
async function getDeployment(id: string): Promise<Deployment>
// Calls apiGet("/api/deployments/" + id)

// Deletes a deployment by ID
async function deleteDeployment(id: string): Promise<void>
// Calls apiDelete("/api/deployments/" + id)

// Triggers a redeploy for an existing deployment
async function redeployDeployment(id: string): Promise<Deployment>
// Calls apiPost("/api/deployments/" + id + "/redeploy", {})
```

### 8.3 Why Everything Uses `multipart/form-data`

The backend's `CreateDeployment` handler parses all requests as multipart form data (via `request.ParseMultipartForm`). This is because zip uploads require multipart, and the handler was written to handle both zip and GitHub deploys through the same multipart parser rather than switching between JSON and multipart based on source type.

The frontend therefore sends ALL create deployment requests as `FormData`, even GitHub deploys that have no file. This is intentional and matches the backend's existing implementation. Do NOT send GitHub deploy requests as JSON — the backend will not parse them correctly.

---

## 9. TypeScript Types

### 9.1 Deployment Type (`types/deployment.ts`)

Mirrors the backend's JSON response shape exactly.

```typescript
type DeploymentStatus = "deploying" | "live" | "failed";
type SourceType = "zip" | "github";

interface Deployment {
  id: string;
  slug: string;
  name: string;
  source_type: SourceType;
  github_url?: string;
  branch: string;
  build_command: string;
  output_directory: string;
  environment_variables?: string; // JSON-encoded string, not parsed object
  status: DeploymentStatus;
  url?: string;
  webhook_secret?: string;
  auto_deploy: boolean;
  created_at: string; // ISO 8601 timestamp
  updated_at: string; // ISO 8601 timestamp
}
```

### 9.2 Preset Type

```typescript
interface DeployPreset {
  id: string;               // unique key for React rendering
  name: string;              // display name on card
  description: string;       // one-line description on card
  icon: string;              // emoji or icon identifier
  githubUrl: string;
  branch: string;
  buildCommand: string;
  outputDirectory: string;
  requiresTextInput: boolean; // true only for "Your Message" preset
  environmentVariables?: Record<string, string>;
}
```

### 9.3 Active Deployment Session Type

```typescript
interface ActiveDeploymentSession {
  id: string;    // deployment UUID
  slug: string;  // deployment slug for URL routing
}
```

Stored in `localStorage` as JSON string under key `corvus_active_deployment`.

---

## 10. Custom Hooks

### 10.1 `useActiveDeployment`

Manages the session state of the user's single active deployment.

**Responsibilities:**
- Read `corvus_active_deployment` from `localStorage` on mount
- Provide the current active deployment session (or null)
- Provide a `setActiveDeployment(id, slug)` function to store after creation
- Provide a `clearActiveDeployment()` function to remove after delete/expiry
- Provide a `hasActiveDeployment` boolean

**Implementation notes:**
- Uses `useState` initialized from `localStorage`
- `setActiveDeployment` writes to both state and `localStorage`
- `clearActiveDeployment` removes from both state and `localStorage`
- Does NOT poll the backend — that is the responsibility of the component or `useDeploymentPolling`

### 10.2 `useDeploymentPolling`

Polls a deployment's status at a regular interval.

**Parameters:**
- `deploymentId: string | null` — the ID to poll (null means do not poll)
- `intervalMs: number` — poll interval in milliseconds (default: 2000)
- `enabled: boolean` — whether polling is active (default: true)

**Returns:**
- `deployment: Deployment | null` — the latest deployment data
- `isLoading: boolean` — true during the first fetch
- `error: ApiError | null` — set if the API returns an error
- `isNotFound: boolean` — true if the API returned 404

**Behavior:**
- On mount (if `deploymentId` is not null and `enabled` is true), immediately fetch once
- Then set an interval to fetch every `intervalMs`
- Stop polling when:
    - `deploymentId` becomes null
    - `enabled` becomes false
    - `deployment.status` is `"live"` or `"failed"` (terminal states)
    - The API returns 404
- Clean up the interval on unmount
- Uses `useEffect` with `deploymentId`, `intervalMs`, and `enabled` as dependencies

### 10.3 `useCountdown`

Calculates and updates a countdown timer every second.

**Parameters:**
- `expiresAt: Date | null` — the expiration timestamp (null means no countdown)

**Returns:**
- `timeRemaining: { minutes: number, seconds: number } | null` — null if expired or no expiration set
- `isExpired: boolean`
- `formattedTime: string` — "MM:SS" format

**Behavior:**
- Uses `useState` + `useEffect` with a 1-second `setInterval`
- Calculates remaining time as `expiresAt.getTime() - Date.now()`
- When remaining time hits 0 or below, sets `isExpired` to true and stops the interval
- The `expiresAt` is calculated by the component: `new Date(deployment.created_at).getTime() + TTL_MS`
- `TTL_MS` comes from `config/constants.ts`

### 10.4 `useFriendCode`

Manages the friend code in `localStorage`.

**Returns:**
- `friendCode: string | null` — the stored friend code
- `setFriendCode(code: string): void` — stores the code
- `clearFriendCode(): void` — removes the code
- `hasFriendCode: boolean`

**Implementation:** Simple `useState` backed by `localStorage` key `corvus_friend_code`.

---

## 11. Configuration Constants (`config/constants.ts`)

All magic values live here. No hardcoded strings or numbers in components.

```typescript
// API
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";

// Deployment TTL
export const DEFAULT_TTL_MINUTES = 15;
export const DEFAULT_TTL_MS = DEFAULT_TTL_MINUTES * 60 * 1000;
export const EXTENDED_TTL_MINUTES = 60; // friend code TTL (or whatever the backend sets)
export const EXTENDED_TTL_MS = EXTENDED_TTL_MINUTES * 60 * 1000;

// Polling
export const POLL_INTERVAL_MS = 2000;     // 2 seconds between polls
export const POLL_TIMEOUT_MS = 120000;    // 2 minutes max polling before showing timeout

// Upload limits
export const MAX_FILE_SIZE_BYTES = 50 * 1024 * 1024; // 50MB
export const MAX_FILE_SIZE_DISPLAY = "50MB";

// "Your Message" preset
export const MAX_MESSAGE_LENGTH = 100;

// localStorage keys
export const STORAGE_KEY_ACTIVE_DEPLOYMENT = "corvus_active_deployment";
export const STORAGE_KEY_FRIEND_CODE = "corvus_friend_code";

// Progress step timing (for simulated steps)
export const STEP_DELAY_SOURCE_RECEIVED_MS = 2000;
export const STEP_DELAY_BUILDING_MS = 4000;
export const STEP_DELAY_STARTING_MS = 6000;

// Preset configurations
export const DEPLOY_PRESETS: DeployPreset[] = [
  {
    id: "vite-starter",
    name: "Vite Starter",
    description: "A minimal Vite app. Deploys in seconds.",
    icon: "⚡",
    githubUrl: "https://github.com/sasta-kro/corvus-preset-vite-starter.git",
    branch: "main",
    buildCommand: "npm ci && npm run build",
    outputDirectory: "dist",
    requiresTextInput: false,
  },
  {
    id: "react-app",
    name: "React App",
    description: "A React + Vite template.",
    icon: "⚛️",
    githubUrl: "https://github.com/sasta-kro/corvus-preset-react-app.git",
    branch: "main",
    buildCommand: "npm ci && npm run build",
    outputDirectory: "dist",
    requiresTextInput: false,
  },
  {
    id: "about-corvus",
    name: "About Corvus",
    description: "A custom page about this platform.",
    icon: "🐦",
    githubUrl: "https://github.com/sasta-kro/corvus-preset-about.git",
    branch: "main",
    buildCommand: "npm ci && npm run build",
    outputDirectory: "dist",
    requiresTextInput: false,
  },
  {
    id: "your-message",
    name: "Your Message",
    description: "Create a page with your custom message.",
    icon: "✍️",
    githubUrl: "https://github.com/sasta-kro/corvus-preset-message.git",
    branch: "main",
    buildCommand: "npm ci && npm run build",
    outputDirectory: "dist",
    requiresTextInput: true,
  },
];
```

---

## 12. State Machine: Landing Page View States

The landing page has four mutually exclusive view states. Only one is rendered at a time in the main content area (below the hero).

```
                    ┌─────────────┐
    Page Load ─────>│  CHECKING   │ (checking localStorage + API)
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
              ▼            ▼            ▼
     ┌────────────┐ ┌───────────┐ ┌──────────┐
     │   DEPLOY   │ │ PROGRESS  │ │  ACTIVE   │
     │   PANEL    │ │   VIEW    │ │DEPLOYMENT │
     └────────────┘ └───────────┘ └──────────┘
              │            │            │
              │            ▼            │
              │      ┌───────────┐      │
              │      │  ACTIVE   │──────┘
              │      │DEPLOYMENT │
              │      └───────────┘
              │            │
              │            ▼
              │      ┌───────────┐
              └──────│  DEPLOY   │ (after delete/expiry)
                     │   PANEL   │
                     └───────────┘
```

| State | Condition | What is rendered |
|---|---|---|
| `CHECKING` | Initial load, localStorage has a deployment ID, waiting for API response | Loading spinner or skeleton |
| `DEPLOY_PANEL` | No active deployment. User can choose a preset, upload zip, or paste GitHub URL | Deploy panel with tabs |
| `PROGRESS` | Deployment was just created, status is `"deploying"` | Progress steps with polling |
| `ACTIVE_DEPLOYMENT` | Deployment exists and status is `"live"` | Live URL, countdown, action buttons |

Transitions:
- `CHECKING` → `DEPLOY_PANEL`: localStorage is empty, or API returned 404, or status is `"failed"`
- `CHECKING` → `PROGRESS`: API returned status `"deploying"`
- `CHECKING` → `ACTIVE_DEPLOYMENT`: API returned status `"live"`
- `DEPLOY_PANEL` → `PROGRESS`: user clicked deploy (any method)
- `PROGRESS` → `ACTIVE_DEPLOYMENT`: polled status changed to `"live"`
- `PROGRESS` → `DEPLOY_PANEL`: polled status changed to `"failed"` (after user clicks "Try Again")
- `ACTIVE_DEPLOYMENT` → `DEPLOY_PANEL`: user clicked delete, or countdown expired and API confirmed gone

The state is managed by a single `useState<"checking" | "deploy" | "progress" | "active">` in `LandingPage.tsx`. The deployment data is held in a separate `useState<Deployment | null>`.

---

## 13. Detailed Component Specifications

### 13.1 `DeployPanel.tsx`

**Role:** Container component that manages which tab is active and handles the transition from tab content to the deploy action.

**State:**
- `activeTab: "quick" | "zip" | "github"` — which tab is selected
- `isDeploying: boolean` — disables all inputs and buttons during API call

**Props:**
- `onDeployStarted(deployment: Deployment): void` — called after the create API returns successfully. The parent (`LandingPage`) uses this to transition to the progress view and store the deployment in `localStorage`.

**Behavior:**
- Renders the `DeployTabs` component with tab switching logic
- Each tab component receives the `isDeploying` state and the `onDeployStarted` callback
- When any tab triggers a deploy:
    1. Sets `isDeploying` to true
    2. Calls the appropriate API function
    3. On success: calls `onDeployStarted(deployment)`
    4. On error: sets `isDeploying` to false, shows error toast
    5. Friend code from `useFriendCode` is passed to every API call

### 13.2 `PresetCard.tsx`

**Props:**
- `preset: DeployPreset`
- `onDeploy(preset: DeployPreset, message?: string): void`
- `disabled: boolean`

**Behavior:**
- Renders a card with the preset's icon, name, and description
- On click:
    - If `preset.requiresTextInput` is true: opens `MessageInputModal`
    - Otherwise: calls `onDeploy(preset)` immediately
- Visual feedback: hover state (border color change or slight scale), disabled state (opacity reduction, no pointer events)

### 13.3 `MessageInputModal.tsx`

**Props:**
- `isOpen: boolean`
- `onClose(): void`
- `onSubmit(message: string): void`
- `maxLength: number` (from constants)

**State:**
- `message: string` — the user's input text

**Behavior:**
- Renders a modal/dialog overlay when `isOpen` is true
- Contains a text input with character counter (`42 / 100`)
- Submit button disabled when message is empty
- On submit: calls `onSubmit(message)`, closes the modal
- On close: clears the message and closes

### 13.4 `DragDropZone.tsx`

**Props:**
- `onFileSelected(file: File): void`
- `onFileRemoved(): void`
- `selectedFile: File | null`
- `error: string | null` — validation error message
- `disabled: boolean`

**State:**
- `isDragOver: boolean` — whether a file is being dragged over the zone

**Behavior:**
- Renders a dashed-border rectangle
- Handles `onDragOver`, `onDragEnter`, `onDragLeave`, `onDrop` events
- On drop:
    - Checks the file extension is `.zip`
    - Checks the file size against `MAX_FILE_SIZE_BYTES`
    - If valid: calls `onFileSelected(file)`
    - If invalid: sets the `error` prop via parent callback
- Contains a hidden `<input type="file" accept=".zip">` triggered by clicking the zone
- When a file is selected: displays file name, formatted file size (e.g., "2.4 MB"), and a remove button (X)
- When a file is removed: calls `onFileRemoved()`

### 13.5 `DeployProgressView.tsx`

**Props:**
- `deployment: Deployment` — the deployment being tracked
- `onComplete(deployment: Deployment): void` — called when status becomes "live"
- `onFailed(): void` — called when status becomes "failed" and user clicks "Try Again"
- `onCancel(): void` — called when user clicks "Cancel" on timeout

**Internal state:**
- `currentStepIndex: number` — which step is "in progress" (0-4)
- `stepStatuses: ("completed" | "in-progress" | "pending" | "failed")[]`
- `elapsedMs: number` — time since deploy started
- `isTimedOut: boolean`

**Uses:** `useDeploymentPolling` to poll the deployment status.

**Step advancement logic (inside a `useEffect`):**
1. On mount: step 0 ("Deployment created") is immediately completed, step 1 is in-progress
2. After `STEP_DELAY_SOURCE_RECEIVED_MS` (2s): step 1 completed, step 2 in-progress
3. After `STEP_DELAY_BUILDING_MS` (4s): step 2 stays in-progress (this is the longest step)
4. After `STEP_DELAY_STARTING_MS` (6s): step 3 in-progress (if status is still deploying)
5. When polled status is `"live"`: all steps completed, call `onComplete` after 500ms delay
6. When polled status is `"failed"`: current in-progress step gets failed state
7. After `POLL_TIMEOUT_MS` (120s): set `isTimedOut` to true, show timeout message

### 13.6 `ProgressStep.tsx`

**Props:**
- `label: string`
- `status: "completed" | "in-progress" | "pending" | "failed"`

**Renders:**
- A single row with an icon on the left and the label text on the right
- Icon by status:
    - `completed`: checkmark (✓) in muted color
    - `in-progress`: spinner or pulsing dot
    - `pending`: empty circle (○) in muted color
    - `failed`: X mark in red
- Text styling by status:
    - `completed`: muted/gray text
    - `in-progress`: normal weight text
    - `pending`: muted/gray text
    - `failed`: red text

### 13.7 `CountdownTimer.tsx`

**Props:**
- `expiresAt: Date`
- `onExpired(): void`

**Uses:** `useCountdown` hook internally.

**Renders:**
- "Expires in MM:SS" when time remaining > 0
- Text turns red/warning color when under 2 minutes remaining
- When expired: calls `onExpired()` and displays "Expired"

### 13.8 `StatusBadge.tsx`

**Props:**
- `status: DeploymentStatus`

**Renders:**
- A small pill/badge component
- `"deploying"`: gray/neutral background, "Deploying..." text
- `"live"`: green/success background, "Live" text
- `"failed"`: red/error background, "Failed" text

### 13.9 `LiveUrlDisplay.tsx`

**Props:**
- `url: string`

**State:**
- `copied: boolean` — toggles the "Copied!" text

**Renders:**
- The URL as a clickable link (opens in new tab)
- A "Copy" button next to it
- On copy click: calls `navigator.clipboard.writeText(url)`, sets `copied` to true, resets after 2 seconds

### 13.10 `DeleteConfirmDialog.tsx`

**Props:**
- `isOpen: boolean`
- `onConfirm(): void`
- `onCancel(): void`
- `isDeleting: boolean` — shows loading state on confirm button

**Renders:**
- A modal/dialog overlay
- Text: "Are you sure you want to delete this deployment? This action cannot be undone."
- Two buttons: "Cancel" and "Delete" (red/destructive styling)
- "Delete" button shows a spinner and is disabled when `isDeleting` is true

### 13.11 `Toast.tsx`

**A simple notification component.** Can use shadcn/ui's Toaster or a minimal custom implementation.

**Types:**
- `success`: green/neutral styling
- `error`: red styling
- `info`: neutral styling

**Behavior:**
- Appears at the top-right or bottom-right of the viewport
- Auto-dismisses after 4 seconds
- Can be dismissed manually by clicking X
- Managed via a context provider or a simple global state

---

## 14. Utility Functions (`lib/utils.ts`)

```typescript
// Formats bytes into human-readable string
// formatFileSize(2621440) → "2.5 MB"
function formatFileSize(bytes: number): string

// Extracts a deployment name from a zip filename
// extractNameFromFilename("my-cool-site.zip") → "my-cool-site"
function extractNameFromFilename(filename: string): string

// Extracts a deployment name from a GitHub URL
// extractNameFromGithubUrl("https://github.com/user/my-app.git") → "my-app"
// extractNameFromGithubUrl("https://github.com/user/my-app") → "my-app"
function extractNameFromGithubUrl(url: string): string

// Ensures a GitHub URL ends with .git
// normalizeGithubUrl("https://github.com/user/repo") → "https://github.com/user/repo.git"
// normalizeGithubUrl("https://github.com/user/repo.git") → "https://github.com/user/repo.git"
function normalizeGithubUrl(url: string): string

// Validates that a string is a valid public GitHub repo URL
// isValidGithubUrl("https://github.com/user/repo") → true
// isValidGithubUrl("https://gitlab.com/user/repo") → false
// isValidGithubUrl("not a url") → false
function isValidGithubUrl(url: string): boolean

// Formats a countdown time
// formatCountdown({ minutes: 12, seconds: 5 }) → "12:05"
function formatCountdown(minutes: number, seconds: number): string

// Formats an ISO timestamp to a human-readable relative or absolute time
// formatTimestamp("2026-03-01T11:11:56.362Z") → "Mar 1, 2026, 11:11 AM" or "2 minutes ago"
function formatTimestamp(isoString: string): string
```

---

## 15. Environment Variables

```
VITE_API_BASE_URL=https://api.corvus.sasta.dev
```

This is the only environment variable the frontend needs. Set via:
- `.env.development`: `VITE_API_BASE_URL=http://localhost:8080` (for local dev against the VM)
- `.env.production`: `VITE_API_BASE_URL=https://api.corvus.sasta.dev` (for the deployed frontend)
- Vercel/Netlify environment settings for production builds

Vite exposes env vars prefixed with `VITE_` via `import.meta.env`.

---

## 16. Error Handling Strategy

### 16.1 API Errors

Every API call is wrapped in try/catch. Error handling depends on where the error occurs:

| Context | Error behavior |
|---|---|
| Creating a deployment | Show error toast with the backend's error message. Return to deploy panel. |
| Polling during progress | If 404: deployment was deleted externally, clear localStorage, return to deploy panel. If other error: show toast, continue polling (may be transient). |
| Deleting a deployment | Show error toast. Keep the active deployment view (do not clear localStorage). |
| Redeploying | Show error toast. Keep the current view. |
| Checking active deployment on page load | If 404: silently clear localStorage, show deploy panel. If other error: show deploy panel (assume deployment is gone). |

### 16.2 Client-Side Validation Errors

Shown inline below the relevant input field. Never as toasts (toasts are for server responses and transient notifications).

| Validation | Error message |
|---|---|
| Zip file too large | "File exceeds the 50MB limit." |
| Non-zip file dropped | "Only .zip files are accepted." |
| GitHub URL invalid | "Only public GitHub repositories are supported." |
| Build command empty (GitHub tab) | "Build command is required for GitHub deployments." |
| Message too long (preset #4) | "Message must be 100 characters or fewer." |

### 16.3 Network Errors

If `fetch()` throws (network unreachable, DNS failure, CORS blocked):
- Show a toast: "Could not connect to the server. Please try again."
- Do not crash. The UI remains functional and the user can retry.

---

## 17. Responsive Design

The implementation uses a simple responsive approach that the web developer will refine:

| Breakpoint | Layout |
|---|---|
| Desktop (>= 1024px) | Full layout as described. Preset cards in 4-column grid or 2x2. Deploy panel centered with max-width. |
| Tablet (768px - 1023px) | Preset cards in 2-column grid. Deploy panel full width with padding. |
| Mobile (< 768px) | Preset cards stacked vertically. Tabs may become a dropdown or accordion. Deploy panel full width. Header collapses friend code input into a menu. |

Tailwind responsive prefixes (`sm:`, `md:`, `lg:`) handle all breakpoints. The web developer will tune exact spacing and layouts.

---

## 18. Accessibility Baseline

Minimum accessibility requirements (the web developer may enhance):

- All interactive elements are focusable and operable via keyboard
- Tab order follows visual order
- Buttons have descriptive text (not just icons)
- Form inputs have associated `<label>` elements
- The drag-and-drop zone has a keyboard-accessible file picker fallback (the "click to browse" feature)
- Status badges use text labels, not color alone
- The modal (message input, delete confirm) traps focus and closes on Escape
- Color contrast ratios meet WCAG AA (4.5:1 for normal text) — this is easy with black-and-white theme

---

## 19. Implementation Order

Build in this order. Each step produces a testable, working increment.

### Step 1: Scaffold and API layer
- `npm create vite@latest corvus-frontend -- --template react-ts`
- Install dependencies: `tailwindcss`, `react-router-dom`, `@radix-ui/react-dialog` (for modals)
- Set up Tailwind config
- Create `api/client.ts` and `api/deployments.ts`
- Create `types/deployment.ts`
- Create `config/constants.ts` with all config values
- Create `lib/utils.ts` with utility functions
- **Test:** Call the backend health endpoint from the browser console to verify CORS works

### Step 2: Routing and layout shell
- Create `App.tsx` with React Router
- Create `Header.tsx` with logo placeholder and "Corvus" text
- Create `Footer.tsx` with links
- Create `LandingPage.tsx` shell with hero section
- Create `DeploymentViewerPage.tsx` shell
- **Test:** Navigate between `/` and `/d/test-slug`, see header/footer render

### Step 3: Deploy panel with Quick Deploy tab
- Create `DeployPanel.tsx`, `DeployTabs.tsx`, `QuickDeployTab.tsx`, `PresetCard.tsx`
- Wire preset card clicks to `createGitHubDeployment` API call
- Store deployment in `localStorage` on success
- **Test:** Click a preset card, see the API call succeed, deployment created in backend

### Step 4: Progress view
- Create `DeployProgressView.tsx`, `ProgressStep.tsx`
- Create `useDeploymentPolling.ts` hook
- Wire the landing page state machine: DEPLOY_PANEL → PROGRESS → ACTIVE_DEPLOYMENT
- Implement simulated step advancement with timers
- **Test:** Click a preset, see progress steps animate, see "Live" when backend finishes

### Step 5: Active deployment view
- Create `ActiveDeploymentView.tsx`, `StatusBadge.tsx`, `LiveUrlDisplay.tsx`, `CountdownTimer.tsx`
- Create `useCountdown.ts` hook
- Create `useActiveDeployment.ts` hook
- Wire session enforcement on page load
- **Test:** Deploy, see live view with countdown. Refresh page, see live view restored from localStorage.

### Step 6: Delete and redeploy
- Create `DeleteConfirmDialog.tsx`
- Wire delete button to API, clear localStorage, return to deploy panel
- Wire redeploy button to API, transition back to progress view
- **Test:** Delete a deployment, see panel return. Redeploy, see progress again.

### Step 7: Zip upload tab
- Create `ZipUploadTab.tsx`, `DragDropZone.tsx`
- Implement drag-and-drop with file validation
- Wire to `createZipDeployment` API call
- **Test:** Drag a zip file, see it upload and deploy

### Step 8: GitHub repo tab
- Create `GitHubRepoTab.tsx`
- Implement URL validation
- Wire to `createGitHubDeployment` API call
- **Test:** Paste a GitHub URL, see it clone, build, and deploy

### Step 9: Message input modal (preset #4)
- Create `MessageInputModal.tsx`
- Wire to the "Your Message" preset card
- **Test:** Click preset, enter text, deploy, see the message on the deployed site

### Step 10: Friend code system
- Create `FriendCodeInput.tsx`
- Create `useFriendCode.ts` hook
- Wire to header and to all create deployment calls
- **Test:** Apply a code, deploy, verify extended TTL in backend response

### Step 11: Deployment viewer page
- Build out `DeploymentViewerPage.tsx` with `DeploymentDetailCard.tsx`
- Fetch deployment by slug (requires slug-based lookup or storing ID alongside slug)
- Wire all action buttons
- **Test:** Navigate to `/d/some-slug`, see full deployment details

### Step 12: Error handling and edge cases
- Add toast notification system
- Handle all API errors with user-facing messages
- Handle network failures gracefully
- Handle polling timeout (120 seconds)
- Handle expired deployments (404 on poll)
- **Test:** Turn off backend, see error messages. Let deployment expire, see expiration flow.

### Step 13: Polish for handoff
- Ensure every component has clean, readable JSX
- Ensure all Tailwind classes use neutral black/white/gray values
- Remove any inline styles
- Ensure consistent naming across all files
- Add brief JSDoc comments on each component's purpose and props
- Verify mobile layout is functional (not pretty, just functional)
- **Handoff:** Zip the project and send to the web developer

---

## 20. Dependencies

```json
{
  "dependencies": {
    "react": "^19.0.0",
    "react-dom": "^19.0.0",
    "react-router-dom": "^7.0.0"
  },
  "devDependencies": {
    "@types/react": "^19.0.0",
    "@types/react-dom": "^19.0.0",
    "typescript": "^5.7.0",
    "vite": "^6.0.0",
    "@vitejs/plugin-react": "^4.0.0",
    "tailwindcss": "^4.0.0",
    "autoprefixer": "^10.0.0",
    "postcss": "^8.0.0"
  }
}
```

shadcn/ui components are copied into the project (not installed as a package). Only install the specific Radix primitives needed for each shadcn component when added (e.g., `@radix-ui/react-dialog` for the modal, `@radix-ui/react-tabs` for the tab component).

---

## 21. Deployment Viewer Page — Slug vs UUID Lookup

The deployment viewer page uses the slug in the URL (`/d/:slug`) for readability and shareability. However, the backend currently only supports lookup by UUID (`GET /api/deployments/:uuid`).

Two approaches:

**Option A: Store both slug and UUID in localStorage and in the URL state.** The frontend always has the UUID available from the create response. When navigating to `/d/:slug`, the frontend looks up the UUID from localStorage and calls the API with the UUID. If the user navigates directly to `/d/:slug` without localStorage (e.g., shared link), the frontend cannot resolve the UUID and shows a "Deployment not found" message.

**Option B: Add slug-based lookup to the backend.** Add a `GET /api/deployments/by-slug/:slug` endpoint (or make the existing endpoint accept both UUID and slug). This allows the deployment viewer page to work for shared links without localStorage.

**Recommendation: Start with Option A.** It requires zero backend changes and works for the primary use case (the user who just deployed). If shared links become important, add the backend endpoint later.

For Option A, the URL for the viewer page should include the UUID as a query parameter or use the UUID directly: `/d/:id`. This changes the route to `/d/:id` where `:id` is the UUID. The slug can be displayed on the page but is not used for the API lookup.

**Updated route:** `/d/:id` where `:id` is the deployment UUID.

---

## 22. What This Document Does NOT Cover

The following are explicitly out of scope for the frontend implementation and are left for the web developer or future versions:

- Final color palette and theme
- Font selection and typography hierarchy
- Animation library integration (Framer Motion or CSS animations)
- Background effects (gradient mesh, particles)
- Logo design and integration
- Micro-interactions and hover effects
- Mobile-specific layout optimizations beyond basic Tailwind responsive classes
- Dark/light mode toggle (implementation is dark-mode-ready via neutral Tailwind classes)
- SEO meta tags and Open Graph images
- Analytics integration
- Performance optimization beyond standard Vite defaults
- End-to-end testing
- Unit testing