# OSA Build - Social Profile System Architecture

**Document Version:** 1.0
**Date:** 2026-01-18
**Status:** Planning Phase
**Related Feature:** OSA Build Onboarding (Screen 2 - Username Selection)

---

## Executive Summary

This document outlines the architecture for BusinessOS's social profile system, enabling users to have unique usernames, public profiles, and the ability to share their AI-generated apps with the community. This builds on the existing workspace/sharing infrastructure and extends it with social discovery features.

---

## Current State Analysis

### Existing User System

**User Table (Better Auth Managed):**
```sql
CREATE TABLE "user" (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255),
    email VARCHAR(255) UNIQUE,
    email_verified BOOLEAN DEFAULT FALSE,
    image VARCHAR(500),
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);
```

**Key Findings:**
- ✅ Basic user identity exists (id, name, email)
- ✅ Profile image support (`image` field)
- ❌ No username field (users only have display name)
- ❌ No public profile system
- ❌ No social graph (follows, likes, etc.)

### Existing Sharing System

BusinessOS already has extensive sharing capabilities:

**1. Contexts (Documents/Profiles) - Lines 32-60**
```sql
CREATE TABLE contexts (
    ...
    is_public BOOLEAN DEFAULT FALSE,
    share_id VARCHAR(32) UNIQUE,  -- Anonymous sharing token
    visibility VARCHAR(20),
    ...
);
```

**2. Projects - Lines 95-120**
```sql
CREATE TABLE projects (
    ...
    visibility VARCHAR(20) DEFAULT 'private',  -- 'private', 'workspace', 'public'
    owner_id VARCHAR(255),
    ...
);
```

**3. Workspaces - Lines 3770-3848**
```sql
CREATE TABLE workspaces (
    ...
    slug VARCHAR(100) UNIQUE,  -- URL-friendly identifier
    owner_id VARCHAR(255),
    ...
);
```

**4. User Workspace Profiles - Lines 3886-3909**
```sql
CREATE TABLE user_workspace_profiles (
    ...
    display_name VARCHAR(255),
    title VARCHAR(255),
    avatar_url VARCHAR(500),
    bio TEXT,
    ...
);
```

**5. Dashboard Sharing - Lines 3934-3997**
```sql
CREATE TABLE user_dashboards (
    ...
    visibility VARCHAR(20) DEFAULT 'private',
    share_token VARCHAR(100) UNIQUE,
    ...
);
```

---

## Proposed Architecture

### Phase 1: Username System

#### 1.1 Database Schema

**New Table: `public_profiles`**
```sql
CREATE TABLE public_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL UNIQUE REFERENCES "user"(id) ON DELETE CASCADE,

    -- Username (unique across platform)
    username VARCHAR(50) NOT NULL UNIQUE,

    -- Public profile info
    display_name VARCHAR(255),
    bio TEXT,
    avatar_url VARCHAR(500),
    location VARCHAR(255),
    website VARCHAR(500),

    -- Social stats
    followers_count INTEGER DEFAULT 0,
    following_count INTEGER DEFAULT 0,
    apps_count INTEGER DEFAULT 0,
    likes_received_count INTEGER DEFAULT 0,

    -- Visibility settings
    is_public BOOLEAN DEFAULT TRUE,
    show_email BOOLEAN DEFAULT FALSE,
    show_location BOOLEAN DEFAULT FALSE,

    -- Profile customization
    theme_color VARCHAR(7) DEFAULT '#667eea',  -- Hex color
    cover_image_url VARCHAR(500),

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_public_profiles_username ON public_profiles(LOWER(username));
CREATE INDEX idx_public_profiles_user_id ON public_profiles(user_id);
CREATE INDEX idx_public_profiles_is_public ON public_profiles(is_public);

COMMENT ON TABLE public_profiles IS 'Public user profiles with unique usernames for social features';
COMMENT ON COLUMN public_profiles.username IS 'Unique username for profile URL (e.g., @johndoe)';
```

**Username Constraints:**
- 3-50 characters
- Alphanumeric + underscores/hyphens only
- Case-insensitive uniqueness
- Cannot start with underscore or hyphen
- Reserved usernames: admin, osa, support, api, app, root, system

**Migration Strategy:**
1. Create `public_profiles` table
2. Backfill existing users with auto-generated usernames (name slug + unique suffix)
3. Add onboarding step for users to customize username
4. Future: Username change (with cooldown period)

#### 1.2 API Endpoints

**Profile Management:**
```
POST   /api/profile/claim-username        # Claim username during onboarding
GET    /api/profile/@{username}           # Get public profile by username
GET    /api/profile/me                    # Get own profile
PATCH  /api/profile/me                    # Update own profile
GET    /api/profile/check-username        # Check username availability
```

**Username Validation Service:**
```go
// internal/services/username_service.go
type UsernameService struct {
    db *pgxpool.Pool
}

func (s *UsernameService) ValidateUsername(username string) error {
    // Check length (3-50)
    // Check format (alphanumeric, _, -)
    // Check reserved words
    // Check profanity filter
    return nil
}

func (s *UsernameService) IsAvailable(username string) (bool, error) {
    // Case-insensitive check
    return true, nil
}

func (s *UsernameService) ClaimUsername(userID, username string) error {
    // Transaction: insert into public_profiles
    return nil
}
```

---

### Phase 2: App Sharing

#### 2.1 Shared Apps System

**Extend Existing Tables:**

**Option A: Use Existing `contexts` Table**
- ✅ Already has `is_public` and `share_id`
- ✅ Supports JSONB `structured_data`
- ❌ Generic - not app-specific

**Option B: New `shared_apps` Table**
```sql
CREATE TABLE shared_apps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,

    -- App identity
    title VARCHAR(255) NOT NULL,
    description TEXT,
    app_type VARCHAR(100),  -- 'starter_app', 'custom_app', 'template'

    -- App configuration (from OSA Build)
    config JSONB NOT NULL,  -- Full app definition
    /*
    {
        "title": "SF Founder Weekend",
        "description": "...",
        "data_sources": [...],
        "ui_config": {...},
        "reasoning": "why this app was created"
    }
    */

    -- Visibility
    visibility VARCHAR(20) DEFAULT 'private',  -- 'private', 'unlisted', 'public'
    share_token VARCHAR(100) UNIQUE,  -- For unlisted sharing

    -- Social engagement
    views_count INTEGER DEFAULT 0,
    likes_count INTEGER DEFAULT 0,
    remixes_count INTEGER DEFAULT 0,  -- How many times copied/forked

    -- Discovery
    tags TEXT[] DEFAULT '{}',
    category VARCHAR(100),

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    published_at TIMESTAMPTZ,  -- When made public

    -- Original app tracking
    remixed_from UUID REFERENCES shared_apps(id) ON DELETE SET NULL
);

CREATE INDEX idx_shared_apps_user_id ON shared_apps(user_id);
CREATE INDEX idx_shared_apps_visibility ON shared_apps(visibility);
CREATE INDEX idx_shared_apps_share_token ON shared_apps(share_token);
CREATE INDEX idx_shared_apps_category ON shared_apps(category);
CREATE INDEX idx_shared_apps_tags ON shared_apps USING GIN(tags);
CREATE INDEX idx_shared_apps_published ON shared_apps(published_at DESC) WHERE visibility = 'public';

COMMENT ON TABLE shared_apps IS 'User-created apps shared with the community';
COMMENT ON COLUMN shared_apps.visibility IS 'private=owner only, unlisted=link only, public=discoverable';
```

**Recommendation:** Use **Option B (new `shared_apps` table)** for:
- Clear separation of concerns
- App-specific fields (views, likes, remixes)
- Better query performance for discovery
- Room for future features (featured apps, app store, etc.)

#### 2.2 App Sharing Workflow

**During OSA Build Onboarding:**
1. User completes personalized apps (Screen 12)
2. Screen 13: "Your OS is ready!"
3. **New Screen 14 (Optional):** "Share Your Apps?"
   - Toggle: "Make my apps discoverable"
   - Privacy options: Private / Unlisted / Public
   - Default: Private

**Post-Onboarding Sharing:**
```
User Dashboard → My Apps → [App Card] → Share Button
    ↓
Share Modal:
    - Visibility selector (Private/Unlisted/Public)
    - Copy share link
    - Publish to community
    - Add tags/category
```

#### 2.3 API Endpoints

**App Sharing:**
```
POST   /api/apps                           # Create/save app
GET    /api/apps                           # List user's apps
GET    /api/apps/{id}                      # Get app details
PATCH  /api/apps/{id}                      # Update app
DELETE /api/apps/{id}                      # Delete app
POST   /api/apps/{id}/publish              # Make public
POST   /api/apps/{id}/unpublish            # Make private
POST   /api/apps/{id}/remix                # Fork/copy app
```

---

### Phase 3: Social Graph

#### 3.1 Follows System

**Table: `user_follows`**
```sql
CREATE TABLE user_follows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    follower_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    following_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(follower_id, following_id),
    CHECK(follower_id != following_id)  -- Can't follow yourself
);

CREATE INDEX idx_user_follows_follower ON user_follows(follower_id);
CREATE INDEX idx_user_follows_following ON user_follows(following_id);

COMMENT ON TABLE user_follows IS 'User follow relationships (Twitter-style)';
```

**API Endpoints:**
```
POST   /api/profile/@{username}/follow     # Follow user
DELETE /api/profile/@{username}/unfollow   # Unfollow user
GET    /api/profile/@{username}/followers  # List followers
GET    /api/profile/@{username}/following  # List following
GET    /api/profile/me/feed                # Activity feed from followed users
```

#### 3.2 Likes System

**Table: `app_likes`**
```sql
CREATE TABLE app_likes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id UUID NOT NULL REFERENCES shared_apps(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(app_id, user_id)
);

CREATE INDEX idx_app_likes_app ON app_likes(app_id);
CREATE INDEX idx_app_likes_user ON app_likes(user_id);

COMMENT ON TABLE app_likes IS 'User likes on shared apps';
```

**Counters Update:**
- Use triggers or application-level increments
- Cached counts in `shared_apps.likes_count`
- Eventual consistency acceptable

#### 3.3 Comments System

**Table: `app_comments`**
```sql
CREATE TABLE app_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id UUID NOT NULL REFERENCES shared_apps(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,

    -- Comment content
    content TEXT NOT NULL,

    -- Threading
    parent_id UUID REFERENCES app_comments(id) ON DELETE CASCADE,

    -- Moderation
    is_edited BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_app_comments_app ON app_comments(app_id, created_at DESC);
CREATE INDEX idx_app_comments_user ON app_comments(user_id);
CREATE INDEX idx_app_comments_parent ON app_comments(parent_id);

COMMENT ON TABLE app_comments IS 'Comments on shared apps';
```

---

### Phase 4: Discovery & Explore

#### 4.1 Explore Feed

**Table: `app_view_events`** (Analytics)
```sql
CREATE TABLE app_view_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id UUID NOT NULL REFERENCES shared_apps(id) ON DELETE CASCADE,
    viewer_id VARCHAR(255) REFERENCES "user"(id) ON DELETE SET NULL,  -- NULL if anonymous

    -- Context
    source VARCHAR(50),  -- 'profile', 'explore', 'search', 'direct_link'
    referrer VARCHAR(500),

    -- Metadata
    viewed_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_app_view_events_app ON app_view_events(app_id, viewed_at DESC);
CREATE INDEX idx_app_view_events_viewer ON app_view_events(viewer_id);

COMMENT ON TABLE app_view_events IS 'Track app views for analytics and trending';
```

**Explore Endpoints:**
```
GET /api/explore/trending              # Trending apps (views + likes last 7 days)
GET /api/explore/new                   # Recently published apps
GET /api/explore/popular               # Most liked apps all-time
GET /api/explore/categories/{category} # Apps by category
GET /api/explore/tags/{tag}            # Apps by tag
GET /api/search/apps?q=...             # Full-text search
GET /api/search/users?q=...            # Search usernames/display names
```

#### 4.2 Recommendation Engine (Future)

**Table: `user_app_interactions`** (ML Features)
```sql
CREATE TABLE user_app_interactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    app_id UUID NOT NULL REFERENCES shared_apps(id) ON DELETE CASCADE,

    -- Interaction types
    viewed BOOLEAN DEFAULT FALSE,
    liked BOOLEAN DEFAULT FALSE,
    remixed BOOLEAN DEFAULT FALSE,
    shared BOOLEAN DEFAULT FALSE,

    -- Time spent
    total_view_time INTEGER DEFAULT 0,  -- seconds

    -- Metadata
    first_interaction_at TIMESTAMPTZ DEFAULT NOW(),
    last_interaction_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_user_app_interactions_user ON user_app_interactions(user_id);
CREATE INDEX idx_user_app_interactions_app ON user_app_interactions(app_id);
```

**Recommendation Strategy:**
1. Collaborative filtering (users who liked X also liked Y)
2. Content-based (similar tags/categories)
3. Social graph (apps from users you follow)
4. Trending boost (recent engagement)

---

## Integration Points

### OSA Build Onboarding Flow

**Modified Screen Sequence:**

1. **Screen 1-2:** Welcome + Username Selection
   - **NEW:** Username claim via `/api/profile/claim-username`
   - Create `public_profiles` entry

2. **Screen 3-12:** Personalization + App Generation
   - Continue as designed

3. **Screen 13:** Your OS is Ready
   - Current: Navigation to `/window`
   - **NEW:** Option to share apps before entering

4. **Screen 14 (NEW - Optional):** Share Your Apps
   - "Would you like to share your apps with the community?"
   - Toggle: Make apps discoverable
   - Privacy selector for each app
   - Skip button (default: all apps stay private)

### Workspace Integration

**Key Decision:** Personal profile vs Workspace profile

- ✅ **Public profiles** are personal (tied to user, not workspace)
- ✅ **Shared apps** can be from personal or workspace context
- ✅ User can be `@johndoe` across all workspaces
- ⚠️ **Workspace slug** (`/ws/acme-corp`) separate from user profile (`/@johndoe`)

**Workspace Apps:**
```sql
ALTER TABLE shared_apps ADD COLUMN workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE;
```
- Apps can be "Personal" (user-owned) or "Workspace" (workspace-owned)
- Workspace apps show as "Created by @johndoe at Acme Corp"

---

## Privacy & Security

### Privacy Levels

**User Profile:**
- **Public:** Profile visible to everyone, searchable
- **Unlisted:** Profile accessible via direct link, not searchable
- **Private:** Profile only visible to user

**Apps:**
- **Private:** Only owner can see
- **Unlisted:** Anyone with link can view (not in explore feed)
- **Public:** Fully discoverable in explore feed

### Security Considerations

1. **Username Squatting Prevention:**
   - Rate limit username changes (1 per 30 days)
   - Reclaim inactive usernames (no activity in 1 year)
   - Verify email before username claim

2. **Content Moderation:**
   - Profanity filter for usernames
   - Report abuse system
   - Admin dashboard for moderation

3. **Data Privacy:**
   - GDPR compliance (export/delete profile)
   - User can delete all shared apps
   - Anonymize analytics after 90 days

4. **API Rate Limiting:**
   - Follow/unfollow: 100/hour
   - Likes: 200/hour
   - Profile updates: 10/hour
   - Search: 60/minute

---

## Frontend Routes

### Public Profile Pages

```
/@{username}                    # User profile page
/@{username}/apps               # User's public apps
/@{username}/likes              # Apps user has liked
/@{username}/followers          # Followers list
/@{username}/following          # Following list
```

**Component Structure:**
```
src/routes/@[username]/
    +layout.svelte              # Profile layout with header
    +layout.server.ts           # Fetch user profile
    +page.svelte                # Profile overview
    apps/
        +page.svelte            # Public apps grid
    likes/
        +page.svelte            # Liked apps
    followers/
        +page.svelte            # Followers list
    following/
        +page.svelte            # Following list
```

### Explore Pages

```
/explore                        # Explore landing
/explore/trending               # Trending apps
/explore/new                    # New apps
/explore/popular                # Popular apps
/explore/category/{category}    # Category browse
/explore/tag/{tag}              # Tag browse
/search?q=...                   # Search results
```

### App Detail Page

```
/app/{share_token}              # View shared app
/app/{share_token}/remix        # Remix/copy app
```

---

## Migration Plan

### Phase 1: Foundation (Week 1-2)
- [ ] Create `public_profiles` table
- [ ] Implement username validation service
- [ ] Add username claim endpoint
- [ ] Backfill existing users with usernames
- [ ] Update onboarding flow (Screen 2)

### Phase 2: App Sharing (Week 3-4)
- [ ] Create `shared_apps` table
- [ ] Implement app publish/unpublish
- [ ] Add share modal to app cards
- [ ] Build public app detail page (`/app/{token}`)
- [ ] Add analytics tracking (`app_view_events`)

### Phase 3: Social Graph (Week 5-6)
- [ ] Create `user_follows` table
- [ ] Implement follow/unfollow endpoints
- [ ] Build public profile pages (`/@username`)
- [ ] Create followers/following lists
- [ ] Activity feed for followed users

### Phase 4: Discovery (Week 7-8)
- [ ] Create `app_likes`, `app_comments` tables
- [ ] Implement like/unlike, comment system
- [ ] Build explore feed (`/explore`)
- [ ] Add search functionality
- [ ] Trending algorithm

### Phase 5: Polish (Week 9-10)
- [ ] Recommendation engine
- [ ] Moderation tools
- [ ] Analytics dashboard
- [ ] Performance optimization
- [ ] Mobile responsive design

---

## Open Questions

1. **App Ownership Transfer:**
   - Can users transfer app ownership to workspace?
   - Can workspace members publish workspace apps under their profile?

2. **Monetization:**
   - Featured apps (promoted placement)?
   - Premium app templates?
   - App marketplace with pricing?

3. **Notifications:**
   - New follower notifications?
   - Comment reply notifications?
   - Like notifications?
   - Weekly digest of activity?

4. **Verification:**
   - Verified badges for notable users?
   - Workspace verification?

5. **Content Licensing:**
   - Default license for shared apps (MIT, CC-BY, etc.)?
   - Allow users to specify license?

---

## Success Metrics

**User Engagement:**
- % of users with claimed usernames
- % of onboarded apps made public
- Active public profiles (viewed in last 30 days)

**Social Activity:**
- Follows per user (avg)
- Likes per app (avg)
- Comments per app (avg)
- Remix rate

**Discovery:**
- Apps viewed via explore feed
- Search queries per day
- Click-through rate (search → app view)

**Retention:**
- Users who return to explore feed
- Users who engage with social features

---

## Technical Stack

**Backend:**
- Go 1.24.1
- PostgreSQL with pgvector
- Redis (caching, rate limiting)
- SQLC for type-safe queries

**Frontend:**
- SvelteKit
- TypeScript
- Tailwind CSS
- Svelte stores for state

**Infrastructure:**
- GCP Cloud Run (existing)
- Cloud Storage (user uploads)
- Cloud CDN (app assets)

---

## Appendix: Example Data Flow

### Username Claim (Onboarding Screen 2)

**Request:**
```http
POST /api/profile/claim-username
Content-Type: application/json

{
  "username": "johndoe",
  "display_name": "John Doe",
  "bio": "Building cool stuff with OSA"
}
```

**Response:**
```json
{
  "success": true,
  "profile": {
    "id": "uuid",
    "user_id": "user-123",
    "username": "johndoe",
    "display_name": "John Doe",
    "bio": "Building cool stuff with OSA",
    "profile_url": "/@johndoe"
  }
}
```

### Publish App (Post-Onboarding)

**Request:**
```http
POST /api/apps
Content-Type: application/json

{
  "title": "SF Founder Weekend",
  "description": "Connect with founders at weekend events",
  "app_type": "starter_app",
  "config": { ... },
  "visibility": "public",
  "tags": ["networking", "events", "san-francisco"],
  "category": "productivity"
}
```

**Response:**
```json
{
  "success": true,
  "app": {
    "id": "uuid",
    "title": "SF Founder Weekend",
    "visibility": "public",
    "share_url": "/app/xyz123abc",
    "public_url": "/@johndoe/apps#xyz123abc",
    "published_at": "2026-01-18T10:30:00Z"
  }
}
```

---

**End of Document**
