db = db.getSiblingDB('rela');

db.createCollection('users');
db.createCollection('tasks');
db.createCollection('boards');
db.createCollection('workspaces');