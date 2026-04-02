#!/bin/bash

set -e  # Exit immediately if a command exits with a non-zero status

# Define cleanup function
cleanup() {
  echo "Cleaning up temporary directory: $temp_dir"
  rm -rf "$temp_dir"
}

# Ensure cleanup runs on script exit, even on errors or interruptions
trap cleanup EXIT INT TERM

# Step 1: Clone the GitHub repo in a temporary directory
temp_dir=$(mktemp -d)
repo_url="git@github.com:controlplane-com/types-go.git"
echo "Cloning repository $repo_url into $temp_dir"
git clone "$repo_url" "$temp_dir"

# Step 2: Copy all directories (excluding the script itself) to the repo root
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo "Copying files from $script_dir to $temp_dir"

# Exclude the script itself and the temp_dir if it's inside script_dir
for item in "$script_dir"/*; do
    # Skip the script file
    if [[ "$item" == "$script_dir/$(basename "$0")" ]]; then
        continue
    fi

    # Copy files and directories to temp_dir
    cp -r "$item" "$temp_dir/pkg"
done

# Step 3: Replace import statements in .go files recursively within temp_dir
echo "Replacing import statements in .go files within $temp_dir"
find "$temp_dir" -type f -name "*.go" -exec sed -i 's|"gitlab\.com/controlplane/controlplane/go-libs/schema|"github\.com/controlplane-com/types-go/pkg|g' {} +

# Step 4: Commit and push changes
cd "$temp_dir"
echo "Adding changes to git"
git add .

# Check if there are any changes to commit
if git diff --cached --quiet; then
    echo "No changes detected. Nothing to commit."
else
    echo "Committing changes"
    git commit -m "Sync changes from upstream"

    echo "Pushing changes to origin"
    git push origin main  # Change 'main' to the correct branch if necessary
fi
