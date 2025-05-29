 #!/bin/bash

echo "Current directory: $(pwd)"
echo "Directory contents:"
ls -la
echo "Running gateway..."
exec /app/gateway