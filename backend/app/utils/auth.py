from datetime import datetime
from typing import Annotated
from dataclasses import dataclass

from fastapi import Depends, HTTPException, status, Request
from sqlalchemy import select, text
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db


@dataclass
class BetterAuthUser:
    """User from Better Auth's user table."""
    id: str
    name: str
    email: str
    email_verified: bool
    image: str | None
    created_at: datetime
    updated_at: datetime


async def get_current_user(
    request: Request,
    db: Annotated[AsyncSession, Depends(get_db)],
) -> BetterAuthUser:
    """
    Validate Better Auth session from cookie and return the user.

    Better Auth stores session token in a cookie named 'better-auth.session_token'.
    """
    # Get session token from cookie
    session_cookie = request.cookies.get("better-auth.session_token")
    
    # Debug: log all cookies received
    print(f"[AUTH DEBUG] All cookies: {dict(request.cookies)}")
    print(f"[AUTH DEBUG] Session cookie value: {session_cookie}")

    if not session_cookie:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Not authenticated",
        )

    # URL decode in case it's encoded
    from urllib.parse import unquote
    session_cookie = unquote(session_cookie)

    # Better Auth signs cookies with HMAC - format is {token}.{signature}
    # We need to extract just the token part (before the dot)
    if '.' in session_cookie:
        session_token = session_cookie.split('.')[0]
    else:
        session_token = session_cookie

    print(f"[AUTH DEBUG] Looking up token: {session_token}")

    # Look up session in Better Auth's session table
    # The token in the cookie is the full token, but Better Auth stores a hash
    # Actually, Better Auth stores the token directly, let's check
    result = await db.execute(
        text('''
            SELECT s.*, u.id as user_id, u.name, u.email, u."emailVerified", u.image, u."createdAt", u."updatedAt"
            FROM session s
            JOIN "user" u ON s."userId" = u.id
            WHERE s.token = :token AND s."expiresAt" > NOW()
        '''),
        {"token": session_token}
    )
    row = result.fetchone()

    if not row:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid or expired session",
        )

    return BetterAuthUser(
        id=row.user_id,
        name=row.name,
        email=row.email,
        email_verified=row.emailVerified,
        image=row.image,
        created_at=row.createdAt,
        updated_at=row.updatedAt,
    )


CurrentUser = Annotated[BetterAuthUser, Depends(get_current_user)]
