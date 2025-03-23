# Change this to the GPU name you want to monitor
$targetGPU = "{{ .Device }}"
$interval = {{ .Interval }} 

while ($true) {
    # Get GPU Info
    $gpuInfo = Get-WmiObject Win32_VideoController | Where-Object { $_.Name -like $targetGPU }

    # Proceed only if the target GPU is found
    if ($gpuInfo) {
        $gpuUtilization = Get-Counter "\GPU Engine(*)\Utilization Percentage" | Select-Object -ExpandProperty CounterSamples
        $gpuMemory = Get-Counter "\GPU Process Memory(*)\Local Usage" | Select-Object -ExpandProperty CounterSamples

        # Calculate total GPU utilization
        $totalGPUUtil = ($gpuUtilization | Measure-Object -Property CookedValue -Sum).Sum

        # Calculate total GPU memory usage
        $totalGPUMem = ($gpuMemory | Measure-Object -Property CookedValue -Sum).Sum

        # Create JSON object
        $jsonOutput = @{
            Name       = $gpuInfo.Name
            Utilization = [math]::Round($totalGPUUtil, 2)  # Rounded percentage
            MemoryUsage = [math]::Round($totalGPUMem / 1MB, 2)  # Convert bytes to MB
        } | ConvertTo-Json -Depth 3

        # Output JSON
        Write-Output $jsonOutput
    } else {
        Write-Output "{ `"Error`": `"GPU '$targetGPU' not found`" }"
        exit 1
    }

    Start-Sleep -Seconds $interval
}