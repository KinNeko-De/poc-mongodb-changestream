print('Starting MongoDB replica set initialization...');

let alreadyInitialized = false;
try {
  const status = rs.status();
  if (status.ok && status.set === 'rs0') {
    alreadyInitialized = true;
    print('Replica set already initialized.');
  }
} catch (e) {
  print('Replica set not initialized, proceeding with initiation...');
}

if (!alreadyInitialized) {
  rs.initiate({
    _id: 'rs0',
    members: [
      { _id: 0, host: 'host.docker.internal:27017' }
    ]
  });
  print('Replica set initiated.');

  // Wait for replica set to initialize
  let attempts = 30;
  while (attempts > 0) {
    try {
      const status = rs.status();
      if (status.ok && status.members && status.members[0].state === 1) {
        print('Replica set initialized successfully!');
        break;
      }
    } catch (e) {
      print('Error checking replica set status: ' + e);
    }
    print('Waiting for replica set to initialize... (' + attempts + ' attempts left)');
    sleep(1000);
    attempts--;
  }

  if (attempts === 0) {
    print('Failed to initialize replica set after multiple attempts');
    quit(1);
  }
}

// Show replica set status
print('Final replica set status:');
printjson(rs.status());
