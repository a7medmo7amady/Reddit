package com.reddit.clone.repository;

import com.reddit.clone.model.User;
import com.reddit.clone.model.UserFollow;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface UserFollowRepository extends JpaRepository<UserFollow, Long> {
    List<UserFollow> findByFollower(User follower);
    List<UserFollow> findByFollowed(User followed);
    Optional<UserFollow> findByFollowerAndFollowed(User follower, User followed);
    boolean existsByFollowerAndFollowed(User follower, User followed);
}
