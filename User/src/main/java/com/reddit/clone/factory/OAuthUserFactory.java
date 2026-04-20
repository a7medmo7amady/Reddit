package com.reddit.clone.factory;

import com.reddit.clone.model.OAuthProvider;
import com.reddit.clone.model.Role;
import com.reddit.clone.model.User;

public class OAuthUserFactory extends UserFactory {

    private final OAuthProvider provider;
    private final String displayName;
    private final String avatar;

    public OAuthUserFactory(OAuthProvider provider, String displayName, String avatar) {
        this.provider = provider;
        this.displayName = displayName;
        this.avatar = avatar;
    }

    @Override
    public User createUser(String username, String email, String password) {
        return new User(username, email, password, displayName, null, avatar, null, Role.USER, provider);
    }

    @Override
    public User createOAuthUser(String username, String email, OAuthProvider provider) {
        return new User(username, email, null, displayName, null, avatar, null, Role.USER, provider);
    }
}
