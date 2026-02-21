#!/usr/bin/env python3
"""
PhantomProxy v14.0 - Backup Script
Automated backup for production deployments
"""

import os
import sys
import shutil
import tarfile
from datetime import datetime
from pathlib import Path

# Configuration
BACKUP_DIR = Path("./backups")
DATA_DIRS = [
    "./phantom.db",
    "./logs",
    "./configs",
    "./certs"
]
RETENTION_DAYS = 30

def create_backup():
    """Create compressed backup"""
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    backup_file = BACKUP_DIR / f"phantom_backup_{timestamp}.tar.gz"
    
    BACKUP_DIR.mkdir(exist_ok=True)
    
    print(f"Creating backup: {backup_file}")
    
    with tarfile.open(backup_file, "w:gz") as tar:
        for path in DATA_DIRS:
            if os.path.exists(path):
                tar.add(path)
                print(f"  Added: {path}")
    
    print(f"Backup created: {backup_file}")
    return backup_file

def cleanup_old_backups():
    """Remove backups older than retention period"""
    if not BACKUP_DIR.exists():
        return
    
    cutoff = datetime.now().timestamp() - (RETENTION_DAYS * 24 * 60 * 60)
    
    for backup in BACKUP_DIR.glob("phantom_backup_*.tar.gz"):
        if backup.stat().st_mtime < cutoff:
            print(f"Removing old backup: {backup}")
            backup.unlink()

def restore_backup(backup_file):
    """Restore from backup"""
    if not os.path.exists(backup_file):
        print(f"Backup not found: {backup_file}")
        return False
    
    print(f"Restoring from: {backup_file}")
    
    with tarfile.open(backup_file, "r:gz") as tar:
        tar.extractall()
    
    print("Restore complete")
    return True

def main():
    if len(sys.argv) > 1:
        if sys.argv[1] == "backup":
            create_backup()
            cleanup_old_backups()
        elif sys.argv[1] == "restore":
            if len(sys.argv) > 2:
                restore_backup(sys.argv[2])
            else:
                print("Usage: backup.sh restore <backup_file>")
        elif sys.argv[1] == "list":
            print("Available backups:")
            for backup in sorted(BACKUP_DIR.glob("phantom_backup_*.tar.gz")):
                print(f"  {backup.name}")
        else:
            print("Usage: backup.sh [backup|restore|list]")
    else:
        # Default: create backup
        create_backup()
        cleanup_old_backups()

if __name__ == '__main__':
    main()
