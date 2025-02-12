/*
document.getElementById('assessmentForm').addEventListener('submit', function (event) {
    event.preventDefault(); // Prevent the form from reloading the page

    // Get user input values
    const age = parseInt(document.getElementById('age').value, 10);
    const mobility = document.getElementById('mobility').value;
    const healthConditions = document.getElementById('health').value;

    // Determine risk level based on simple logic
    let riskLevel = "Low";

    if (age > 65 || mobility === "severe" || healthConditions.length > 50) {
        riskLevel = "High";
    } else if (age > 50 || mobility === "minor" || healthConditions.length > 20) {
        riskLevel = "Moderate";
    }

    // Display the result
    document.getElementById('result').style.display = 'block';
    document.getElementById('riskLevel').textContent = riskLevel;
});
*/