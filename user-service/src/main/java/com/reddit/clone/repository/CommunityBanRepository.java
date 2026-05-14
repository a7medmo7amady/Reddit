package com.reddit.clone.repository;

import com.reddit.clone.model.CommunityBan;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface CommunityBanRepository extends JpaRepository<CommunityBan, Long> {
    boolean existsByUserIdAndCommunityName(Long userId, String communityName);
    Optional<CommunityBan> findByUserIdAndCommunityName(Long userId, String communityName);
}
