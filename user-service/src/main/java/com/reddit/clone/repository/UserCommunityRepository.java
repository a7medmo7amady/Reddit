package com.reddit.clone.repository;

import com.reddit.clone.model.Community;
import com.reddit.clone.model.User;
import com.reddit.clone.model.UserCommunity;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface UserCommunityRepository extends JpaRepository<UserCommunity, Long> {
    List<UserCommunity> findByUser(User user);
    Optional<UserCommunity> findByUserAndCommunity(User user, Community community);
    boolean existsByUserAndCommunity(User user, Community community);
    long countByCommunity(Community community);
}
