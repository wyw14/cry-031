INSERT INTO communities (id, name, address, description)
VALUES
    ('community-riverside', '滨河社区', '新桥路 18 号', '滨河居民自治社区'),
    ('community-garden', '桂园社区', '桂园路 6 号', '老幼友好社区')
ON CONFLICT (id) DO NOTHING;

INSERT INTO service_areas (id, name)
VALUES ('area-elder', '助老服务'), ('area-environment', '环境维护')
ON CONFLICT (id) DO NOTHING;

-- The aggregate snapshot is seeded by cmd/server using repository.DemoState so it is repeatable
-- and never overwrites an existing volunteer_state row.
