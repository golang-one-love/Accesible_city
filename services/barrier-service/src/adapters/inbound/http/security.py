from functools import lru_cache
from typing import Optional
from uuid import UUID

from fastapi import HTTPException, Header, status
from jose import jwt, JWTError

from src.config import settings

JWT_ALGORITHM = "RS256"


@lru_cache(maxsize=1)
def _public_key() -> str:
    with open(settings.JWT_PUBLIC_KEY_PATH, "r") as f:
        return f.read()


def decode_access_token(token: str) -> dict:
    try:
        claims = jwt.decode(token, _public_key(), algorithms=[JWT_ALGORITHM])
    except JWTError:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid or expired token",
        )
    if claims.get("type") != "access":
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Not an access token",
        )
    return claims


def get_current_user_id(authorization: str = Header(None)) -> UUID:
    if not authorization:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Missing authorization header")
    parts = authorization.split()
    if len(parts) != 2 or parts[0].lower() != "bearer":
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid authorization header format")

    claims = decode_access_token(parts[1])
    try:
        return UUID(claims["sub"])
    except (KeyError, ValueError):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid token subject",
        )


def get_current_user_id_optional(authorization: str = Header(None)) -> Optional[UUID]:
    if not authorization:
        return None
    parts = authorization.split()
    if len(parts) != 2 or parts[0].lower() != "bearer":
        return None
    try:
        claims = decode_access_token(parts[1])
        return UUID(claims["sub"])
    except (JWTError, HTTPException, KeyError, ValueError):
        return None
