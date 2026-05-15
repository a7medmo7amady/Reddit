package com.reddit.clone.service;

import com.reddit.clone.exception.OAuthLoginRequiredException;
import com.reddit.clone.factory.UserFactory;
import com.reddit.clone.model.OAuthProvider;
import com.reddit.clone.model.Role;
import com.reddit.clone.model.User;
import com.reddit.clone.repository.UserRepository;
import com.reddit.clone.security.PasswordService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class AuthServiceOAuthTest {

    @Mock UserRepository userRepository;
    @Mock UserFactory userFactory;
    @Mock PasswordService passwordService;

    @InjectMocks AuthService authService;

    private User makeUser(String username, String email, OAuthProvider provider) {
        return new User(username, email, null, null, null, null, null, Role.USER, provider);
    }

    @BeforeEach
    void mockSave() {
        lenient().when(userRepository.save(any())).thenAnswer(inv -> inv.getArgument(0));
    }

    @Test
    void newUser_createsUserAndSetsDisplayNameAndAvatar() {
        when(userRepository.findByEmail("new@example.com")).thenReturn(Optional.empty());
        when(userRepository.existsByUsername("john_doe")).thenReturn(false);

        User created = makeUser("john_doe", "new@example.com", OAuthProvider.GOOGLE);
        when(userFactory.createOAuthUser("john_doe", "new@example.com", OAuthProvider.GOOGLE)).thenReturn(created);

        User result = authService.findOrCreateOAuthUser("new@example.com", "John Doe", "https://pic.url", OAuthProvider.GOOGLE);

        assertThat(result.getDisplayName()).isEqualTo("John Doe");
        assertThat(result.getAvatar()).isEqualTo("https://pic.url");
        verify(userRepository).save(created);
    }

    @Test
    void newUser_nullName_derivesUsernameFromEmailPrefix() {
        when(userRepository.findByEmail("alice@example.com")).thenReturn(Optional.empty());
        when(userRepository.existsByUsername("alice")).thenReturn(false);

        User created = makeUser("alice", "alice@example.com", OAuthProvider.GOOGLE);
        when(userFactory.createOAuthUser(eq("alice"), eq("alice@example.com"), eq(OAuthProvider.GOOGLE))).thenReturn(created);

        authService.findOrCreateOAuthUser("alice@example.com", null, null, OAuthProvider.GOOGLE);

        verify(userFactory).createOAuthUser("alice", "alice@example.com", OAuthProvider.GOOGLE);
    }

    @Test
    void newUser_usernameCollision_appendsNumericSuffix() {
        when(userRepository.findByEmail("new@example.com")).thenReturn(Optional.empty());
        when(userRepository.existsByUsername("john_doe")).thenReturn(true);
        when(userRepository.existsByUsername("john_doe_1")).thenReturn(false);

        User created = makeUser("john_doe_1", "new@example.com", OAuthProvider.GOOGLE);
        when(userFactory.createOAuthUser(eq("john_doe_1"), any(), any())).thenReturn(created);

        authService.findOrCreateOAuthUser("new@example.com", "John Doe", null, OAuthProvider.GOOGLE);

        verify(userFactory).createOAuthUser("john_doe_1", "new@example.com", OAuthProvider.GOOGLE);
    }

    @Test
    void existingUser_linksOAuthProviderIfNull() {
        User existing = makeUser("john_doe", "john@example.com", null);
        when(userRepository.findByEmail("john@example.com")).thenReturn(Optional.of(existing));

        authService.findOrCreateOAuthUser("john@example.com", "John", "pic", OAuthProvider.GOOGLE);

        assertThat(existing.getOauthProvider()).isEqualTo(OAuthProvider.GOOGLE);
    }

    @Test
    void existingUser_setsDisplayNameIfBlank() {
        User existing = makeUser("john_doe", "john@example.com", OAuthProvider.GOOGLE);
        when(userRepository.findByEmail("john@example.com")).thenReturn(Optional.of(existing));

        authService.findOrCreateOAuthUser("john@example.com", "John", null, OAuthProvider.GOOGLE);

        assertThat(existing.getDisplayName()).isEqualTo("John");
    }

    @Test
    void existingUser_setsAvatarIfBlank() {
        User existing = makeUser("john_doe", "john@example.com", OAuthProvider.GOOGLE);
        when(userRepository.findByEmail("john@example.com")).thenReturn(Optional.of(existing));

        authService.findOrCreateOAuthUser("john@example.com", "John", "https://pic.url", OAuthProvider.GOOGLE);

        assertThat(existing.getAvatar()).isEqualTo("https://pic.url");
    }

    @Test
    void existingUser_doesNotOverwriteExistingDisplayName() {
        User existing = makeUser("john_doe", "john@example.com", OAuthProvider.GOOGLE);
        existing.setDisplayName("Already Set");
        when(userRepository.findByEmail("john@example.com")).thenReturn(Optional.of(existing));

        authService.findOrCreateOAuthUser("john@example.com", "New Name", null, OAuthProvider.GOOGLE);

        assertThat(existing.getDisplayName()).isEqualTo("Already Set");
    }

    @Test
    void existingUser_doesNotOverwriteExistingAvatar() {
        User existing = makeUser("john_doe", "john@example.com", OAuthProvider.GOOGLE);
        existing.setAvatar("https://existing.url");
        when(userRepository.findByEmail("john@example.com")).thenReturn(Optional.of(existing));

        authService.findOrCreateOAuthUser("john@example.com", "John", "https://new.url", OAuthProvider.GOOGLE);

        assertThat(existing.getAvatar()).isEqualTo("https://existing.url");
    }

    @Test
    void bannedUser_throwsSecurityException() {
        User banned = makeUser("john_doe", "john@example.com", OAuthProvider.GOOGLE);
        banned.setBanned(true);
        when(userRepository.findByEmail("john@example.com")).thenReturn(Optional.of(banned));

        assertThatThrownBy(() ->
            authService.findOrCreateOAuthUser("john@example.com", "John", null, OAuthProvider.GOOGLE)
        ).isInstanceOf(SecurityException.class);
    }

    
    @Test
    void login_oauthUser_throwsOAuthLoginRequiredException() {
        User oauthUser = makeUser("john_doe", "john@example.com", OAuthProvider.GOOGLE);
        when(userRepository.findByEmail("john@example.com")).thenReturn(Optional.of(oauthUser));

        assertThatThrownBy(() -> authService.login("john@example.com", "somepassword"))
                .isInstanceOf(OAuthLoginRequiredException.class);
    }
}
