.PHONY: start stop restart status logs clean test help

# Start the MySQL cluster
start:
	@echo "Starting MySQL primary-replica cluster..."
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "✓ Cluster started"
	@$(MAKE) status

# Stop the cluster
stop:
	@echo "Stopping MySQL cluster..."
	docker-compose stop
	@echo "✓ Cluster stopped"

# Restart the cluster
restart: stop start

# Show cluster status
status:
	@echo "=== Cluster Status ==="
	@docker-compose ps
	@echo ""
	@echo "=== Replication Status ==="
	@docker exec plumego-mysql-replica1 mysql -uroot -prootpassword -e "SHOW SLAVE STATUS\G" 2>/dev/null | grep "Slave_.*_Running" || echo "Replica1: Not configured"
	@docker exec plumego-mysql-replica2 mysql -uroot -prootpassword -e "SHOW SLAVE STATUS\G" 2>/dev/null | grep "Slave_.*_Running" || echo "Replica2: Not configured"

# Show logs
logs:
	docker-compose logs -f

# Show replication lag
lag:
	@echo "=== Replication Lag ==="
	@docker exec plumego-mysql-replica1 mysql -uroot -prootpassword -e "SHOW SLAVE STATUS\G" 2>/dev/null | grep "Seconds_Behind_Master" || echo "Replica1: N/A"
	@docker exec plumego-mysql-replica2 mysql -uroot -prootpassword -e "SHOW SLAVE STATUS\G" 2>/dev/null | grep "Seconds_Behind_Master" || echo "Replica2: N/A"

# Connect to primary
primary-shell:
	@docker exec -it plumego-mysql-primary mysql -uroot -prootpassword testdb

# Connect to replica1
replica1-shell:
	@docker exec -it plumego-mysql-replica1 mysql -uroot -prootpassword testdb

# Connect to replica2
replica2-shell:
	@docker exec -it plumego-mysql-replica2 mysql -uroot -prootpassword testdb

# Run integration tests
test:
	@echo "Running integration tests..."
	cd .. && go test -v -tags=integration -timeout 60s

# Clean up everything (including volumes)
clean:
	@echo "⚠ This will delete all data!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker-compose down -v; \
		echo "✓ Cleaned up"; \
	fi

# Show help
help:
	@echo "MySQL Primary-Replica Test Environment"
	@echo ""
	@echo "Usage:"
	@echo "  make start          - Start the cluster"
	@echo "  make stop           - Stop the cluster"
	@echo "  make restart        - Restart the cluster"
	@echo "  make status         - Show cluster status"
	@echo "  make logs           - Show logs (follow mode)"
	@echo "  make lag            - Show replication lag"
	@echo "  make primary-shell  - Connect to primary MySQL"
	@echo "  make replica1-shell - Connect to replica1 MySQL"
	@echo "  make replica2-shell - Connect to replica2 MySQL"
	@echo "  make test           - Run integration tests"
	@echo "  make clean          - Remove everything (including data)"
	@echo "  make help           - Show this help"
	@echo ""
	@echo "Ports:"
	@echo "  Primary:  localhost:3306"
	@echo "  Replica1: localhost:3307"
	@echo "  Replica2: localhost:3308"
	@echo ""
	@echo "Credentials:"
	@echo "  User:     testuser / testpass"
	@echo "  Root:     root / rootpassword"
	@echo "  Database: testdb"
