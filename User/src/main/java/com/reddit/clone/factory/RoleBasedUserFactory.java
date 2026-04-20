package com.reddit.clone.factory;

import com.reddit.clone.model.Role;
import com.reddit.clone.model.User;
import org.springframework.stereotype.Component;

@Component
public class RoleBasedUserFactory extends UserFactory {

    private final Role role;

    public RoleBasedUserFactory() {
        this.role = Role.USER;
    }

    @Override
    public User createUser(String username, String email, String password) {
        return new User(username, email, password, null, null, null, null, role, null);
    }
}
