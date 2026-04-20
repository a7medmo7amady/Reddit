package com.reddit.clone.exception;

public class OAuthLoginRequiredException extends RuntimeException {
    public OAuthLoginRequiredException(String provider) {
        super("This account uses " + provider + " login. Please sign in with OAuth.");
    }
}
