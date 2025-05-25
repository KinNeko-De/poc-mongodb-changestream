// This script runs during MongoDB container initialization
print('Starting MongoDB initialization...');

// Create the application database
db = db.getSiblingDB('changestream_poc');

// Create the test_data collection
db.createCollection('test_data');

print('MongoDB initialization completed successfully!');
