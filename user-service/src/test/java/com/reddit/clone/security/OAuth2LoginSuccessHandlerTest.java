package com.reddit.clone.security;

import com.reddit.clone.model.OAuthProvider;
import com.reddit.clone.model.Role;
import com.reddit.clone.model.User;
import com.reddit.clone.service.AuthService;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.HttpHeaders;
import org.springframework.security.core.Authentication;
import org.springframework.security.oauth2.core.user.OAuth2User;
import org.springframework.test.util.ReflectionTestUtils;

import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class OAuth2LoginSuccessHandlerTest {

    @Mock AuthService authService;
    @Mock JWTservice jwtService;
    @Mock TokenService tokenService;

    @InjectMocks OAuth2LoginSuccessHandler handler;

    @Mock HttpServletRequest request;
    @Mock HttpServletResponse response;
    @Mock Authentication authentication;
    @Mock OAuth2User oauthUser;

    private User testUser;

    @BeforeEach
    void setUp() {
        ReflectionTestUtils.setField(handler, "clientRedirectUrl", "http://localhost:3000");

        testUser = new User("test_user", "user@example.com", null, "Test User",
                null, "https://pic.url", null, Role.USER, OAuthProvider.GOOGLE);

        when(authentication.getPrincipal()).thenReturn(oauthUser);
        when(oauthUser.getAttribute("email")).thenReturn("user@example.com");
        when(oauthUser.getAttribute("name")).thenReturn("Test User");
        when(oauthUser.getAttribute("picture")).thenReturn("https://pic.url");
        when(authService.findOrCreateOAuthUser(
                "user@example.com", "Test User", "https://pic.url", OAuthProvider.GOOGLE))
                .thenReturn(testUser);
        when(tokenService.create(any())).thenReturn("refresh-token-value");
        when(jwtService.generateAccessToken(testUser)).thenReturn("access-token-value");
    }

    @Test
    void setsRefreshTokenCookie() throws Exception {
        handler.onAuthenticationSuccess(request, response, authentication);

        verify(response).addHeader(
                eq(HttpHeaders.SET_COOKIE),
                argThat(v -> v.contains("refresh_token=refresh-token-value")
                        && v.contains("HttpOnly")
                        && v.contains("Path=/auth")));
    }

    @Test
    void redirectsToClientWithAccessToken() throws Exception {
        handler.onAuthenticationSuccess(request, response, authentication);

        verify(response).sendRedirect("http://localhost:3000?accessToken=access-token-value");
    }

    @Test
    void callsFindOrCreateWithCorrectAttributes() throws Exception {
        handler.onAuthenticationSuccess(request, response, authentication);

        verify(authService).findOrCreateOAuthUser(
                "user@example.com", "Test User", "https://pic.url", OAuthProvider.GOOGLE);
    }

    @Test
    void createsRefreshTokenForUser() throws Exception {
        handler.onAuthenticationSuccess(request, response, authentication);

        verify(tokenService).create(testUser.getId());
    }

    @Test
    void generatesAccessTokenForUser() throws Exception {
        handler.onAuthenticationSuccess(request, response, authentication);

        verify(jwtService).generateAccessToken(testUser);
    }
}
