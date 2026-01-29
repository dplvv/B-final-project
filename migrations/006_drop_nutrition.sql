-- +migrate Up

drop table if exists nutrition_diary;
drop table if exists nutrition_water_logs;
drop table if exists nutrition_items;

-- +migrate Down
-- (not implemented)
