#!/bin/bash
set -e

echo "=== Setting up MySQL Replication ==="

# Wait for primary to be ready
echo "Waiting for primary to be ready..."
until mysql -h mysql-primary -uroot -prootpassword -e "SELECT 1" &> /dev/null; do
    echo "Primary not ready yet, waiting..."
    sleep 2
done
echo "✓ Primary is ready"

# Wait for replicas to be ready
echo "Waiting for replica1 to be ready..."
until mysql -h mysql-replica1 -uroot -prootpassword -e "SELECT 1" &> /dev/null; do
    echo "Replica1 not ready yet, waiting..."
    sleep 2
done
echo "✓ Replica1 is ready"

echo "Waiting for replica2 to be ready..."
until mysql -h mysql-replica2 -uroot -prootpassword -e "SELECT 1" &> /dev/null; do
    echo "Replica2 not ready yet, waiting..."
    sleep 2
done
echo "✓ Replica2 is ready"

# Setup replication on replica1
echo "Setting up replication on replica1..."
mysql -h mysql-replica1 -uroot -prootpassword <<EOF
STOP SLAVE;
CHANGE REPLICATION SOURCE TO
    SOURCE_HOST='mysql-primary',
    SOURCE_USER='repl',
    SOURCE_PASSWORD='replpass',
    SOURCE_AUTO_POSITION=1;
START SLAVE;
EOF

# Wait a moment for replication to start
sleep 2

# Check replica1 status
echo "Checking replica1 status..."
REPLICA1_STATUS=$(mysql -h mysql-replica1 -uroot -prootpassword -e "SHOW SLAVE STATUS\G" | grep "Slave_IO_Running" | awk '{print $2}')
if [ "$REPLICA1_STATUS" == "Yes" ]; then
    echo "✓ Replica1 replication is running"
else
    echo "⚠ Replica1 replication status: $REPLICA1_STATUS"
fi

# Setup replication on replica2
echo "Setting up replication on replica2..."
mysql -h mysql-replica2 -uroot -prootpassword <<EOF
STOP SLAVE;
CHANGE REPLICATION SOURCE TO
    SOURCE_HOST='mysql-primary',
    SOURCE_USER='repl',
    SOURCE_PASSWORD='replpass',
    SOURCE_AUTO_POSITION=1;
START SLAVE;
EOF

# Wait a moment for replication to start
sleep 2

# Check replica2 status
echo "Checking replica2 status..."
REPLICA2_STATUS=$(mysql -h mysql-replica2 -uroot -prootpassword -e "SHOW SLAVE STATUS\G" | grep "Slave_IO_Running" | awk '{print $2}')
if [ "$REPLICA2_STATUS" == "Yes" ]; then
    echo "✓ Replica2 replication is running"
else
    echo "⚠ Replica2 replication status: $REPLICA2_STATUS"
fi

echo ""
echo "=== Replication Setup Complete ==="
echo "Primary:  localhost:3306"
echo "Replica1: localhost:3307"
echo "Replica2: localhost:3308"
echo ""
echo "Database: testdb"
echo "User:     testuser"
echo "Password: testpass"
echo ""
echo "Root Password: rootpassword"
