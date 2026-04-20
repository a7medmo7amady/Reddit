package com.reddit.clone.factory;

import com.reddit.clone.model.OAuthProvider;
import com.reddit.clone.model.User;

public abstract class UserFactory {
    public abstract User createUser(String username, String email, String password);
    public abstract User createOAuthUser(String username, String email, OAuthProvider provider);
}
