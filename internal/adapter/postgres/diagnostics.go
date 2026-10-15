package postgres

import "context"

func (store *Store) Diagnostics(ctx context.Context) (map[string]string, error) {
	result := map[string]string{}
	var ready, processes, projections, checkpoints, slots string
	query := `SELECT 'ready',
 (SELECT count(*)::text FROM process_outbox WHERE NOT acknowledged),
 (SELECT count(*)::text FROM outbox),
 (SELECT COALESCE(jsonb_object_agg(consumer,lsn::text),'{}'::jsonb)::text FROM cdc_checkpoints),
 (SELECT COALESCE(jsonb_object_agg(slot_name,jsonb_build_object('active',active,'wal_status',wal_status,'restart_lsn',restart_lsn::text,'confirmed_flush_lsn',confirmed_flush_lsn::text)),'{}'::jsonb)::text FROM pg_replication_slots WHERE slot_name=$1 OR slot_name IN (SELECT consumer FROM cdc_checkpoints))`
	err := store.pool.QueryRow(ctx, query, store.config.ReplicationSlot).Scan(&ready, &processes, &projections, &checkpoints, &slots)
	if err != nil {
		return nil, databaseError(err)
	}

	result["PostgreSQL"] = ready
	result["PendingProcesses"] = processes
	result["ProjectionEvents"] = projections
	result["Checkpoints"] = checkpoints
	result["ReplicationSlots"] = slots

	return result, nil
}
